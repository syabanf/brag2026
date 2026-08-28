package usecase

import (
	"context"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type Tyfcb struct {
	tyfcb   domain.TyfcbRepository
	members domain.MemberRepository
	ledger  domain.LedgerRepository
	seasons domain.SeasonRepository
	tx      domain.TxManager
}

func NewTyfcb(
	tyfcb domain.TyfcbRepository,
	members domain.MemberRepository,
	ledger domain.LedgerRepository,
	seasons domain.SeasonRepository,
	tx domain.TxManager,
) *Tyfcb {
	return &Tyfcb{tyfcb: tyfcb, members: members, ledger: ledger, seasons: seasons, tx: tx}
}

type SubmitTyfcbInput struct {
	// BuyerID is the member who bought, and who therefore earns the points.
	BuyerID string
	// SellerID records the transaction. Defaults to the caller's own member id.
	SellerID string
	Nilai    float64
	Tanggal  time.Time
}

// Submit records a pending TYFCB. Points are computed now but only reach the
// ledger once an admin verifies it.
func (u *Tyfcb) Submit(ctx context.Context, in SubmitTyfcbInput, submittedBy *string) (*domain.TyfcbEntry, error) {
	if in.BuyerID == "" || in.Nilai <= 0 || in.Tanggal.IsZero() {
		return nil, domain.Invalid("buyer_id, nilai, dan tanggal wajib diisi.")
	}
	if in.BuyerID == in.SellerID {
		return nil, domain.Invalid("Tidak bisa mencatat TYFCB ke diri sendiri.")
	}

	seller, err := u.members.FindByID(ctx, in.SellerID)
	if err != nil {
		return nil, err
	}
	if seller == nil {
		return nil, domain.NotFound("Profil member tidak ditemukan di season ini.")
	}

	buyer, err := u.members.FindByID(ctx, in.BuyerID)
	if err != nil {
		return nil, err
	}
	if buyer == nil || buyer.SeasonID != seller.SeasonID {
		return nil, domain.Invalid("Pembeli tidak ditemukan di season ini.")
	}

	// The nth transaction between the same pair is worth 1/n of the band.
	prior, err := u.tyfcb.CountPair(ctx, in.BuyerID, seller.ID, seller.SeasonID)
	if err != nil {
		return nil, err
	}
	pairOrdinal := prior + 1

	const eventMultiplier = 1.0
	score := domain.TyfcbScore(in.Nilai, pairOrdinal, eventMultiplier)
	multiplier := eventMultiplier

	entry := &domain.TyfcbEntry{
		SeasonID:               seller.SeasonID,
		GiverID:                in.BuyerID,
		ReceiverID:             seller.ID,
		Nilai:                  in.Nilai,
		Tanggal:                in.Tanggal,
		Status:                 domain.TyfcbPending,
		ComputedScore:          &score,
		PairOrdinal:            &pairOrdinal,
		EventMultiplierApplied: &multiplier,
	}

	id, err := u.tyfcb.Create(ctx, entry, submittedBy)
	if err != nil {
		return nil, err
	}
	entry.ID = id

	return entry, nil
}

// SetStatus moves an entry between pending/verified/rejected and keeps the
// ledger consistent: crossing into verified credits the points, leaving it
// writes an equal and opposite reversal. Both happen in one transaction with
// the status change so a partial failure cannot skew a leaderboard.
func (u *Tyfcb) SetStatus(ctx context.Context, id string, next domain.TyfcbStatus, actorID string) error {
	if !domain.ValidTyfcbStatus(string(next)) {
		return domain.Invalid("Status tidak valid.")
	}

	entry, err := u.tyfcb.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if entry == nil {
		return domain.NotFound("Entry tidak ditemukan.")
	}
	if entry.Status == next {
		return domain.Invalid("Status sudah sama.")
	}
	if entry.Status == domain.TyfcbVoid {
		return domain.Conflict("Entry sudah di-void dan tidak bisa diubah.")
	}

	giver, err := u.members.FindByID(ctx, entry.GiverID)
	if err != nil {
		return err
	}

	credits := entry.Status != domain.TyfcbVerified && next == domain.TyfcbVerified
	reverses := entry.Status == domain.TyfcbVerified && next != domain.TyfcbVerified

	return u.tx.WithinTx(ctx, func(ctx context.Context) error {
		if entry.ComputedScore != nil && (credits || reverses) {
			points := *entry.ComputedScore
			keterangan := "TYFCB verified"
			if reverses {
				points = -points
				keterangan = "TYFCB reversal"
			}

			var teamID *string
			if giver != nil {
				teamID = giver.TeamID
			}
			ref := entry.ID

			if err := u.ledger.Append(ctx, &domain.LedgerEntry{
				SeasonID:   entry.SeasonID,
				MemberID:   &entry.GiverID,
				TeamID:     teamID,
				Kategori:   domain.CategoryTyfcb,
				Points:     points,
				SumberRef:  &ref,
				Keterangan: &keterangan,
			}); err != nil {
				return err
			}
		}

		var verifiedBy *string
		var verifiedAt *time.Time
		if next == domain.TyfcbVerified {
			now := time.Now()
			verifiedBy, verifiedAt = &actorID, &now
		}

		return u.tyfcb.UpdateStatus(ctx, id, next, verifiedBy, verifiedAt)
	})
}

// Void cancels an entry outright, reversing any points it had earned.
func (u *Tyfcb) Void(ctx context.Context, id, actorID string) error {
	entry, err := u.tyfcb.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if entry == nil {
		return domain.NotFound("Entry tidak ditemukan.")
	}
	if entry.Status == domain.TyfcbVoid {
		return domain.Conflict("Entry sudah di-void.")
	}

	giver, err := u.members.FindByID(ctx, entry.GiverID)
	if err != nil {
		return err
	}

	return u.tx.WithinTx(ctx, func(ctx context.Context) error {
		if entry.Status == domain.TyfcbVerified && entry.ComputedScore != nil {
			var teamID *string
			if giver != nil {
				teamID = giver.TeamID
			}
			ref := entry.ID
			keterangan := "TYFCB void"
			points := -*entry.ComputedScore

			if err := u.ledger.Append(ctx, &domain.LedgerEntry{
				SeasonID:   entry.SeasonID,
				MemberID:   &entry.GiverID,
				TeamID:     teamID,
				Kategori:   domain.CategoryTyfcb,
				Points:     points,
				SumberRef:  &ref,
				Keterangan: &keterangan,
			}); err != nil {
				return err
			}
		}
		return u.tyfcb.Void(ctx, id, actorID)
	})
}

func (u *Tyfcb) List(ctx context.Context, f domain.TyfcbFilter) ([]domain.TyfcbEntry, error) {
	return u.tyfcb.List(ctx, f)
}

func (u *Tyfcb) CountByStatus(ctx context.Context, seasonID string) (map[string]int, error) {
	return u.tyfcb.CountByStatus(ctx, seasonID)
}
