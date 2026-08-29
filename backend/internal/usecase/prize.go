package usecase

import (
	"context"
	"strings"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// Prize runs the two-layer reward system: a pool of items, plus raffle tickets
// that decide who is in the draw for the ones allocated by lottery.
type Prize struct {
	prizes  domain.PrizeRepository
	members domain.MemberRepository
	seasons domain.SeasonRepository
	badges  *Badges
	tx      domain.TxManager
}

func NewPrize(
	prizes domain.PrizeRepository,
	members domain.MemberRepository,
	seasons domain.SeasonRepository,
	badges *Badges,
	tx domain.TxManager,
) *Prize {
	return &Prize{prizes: prizes, members: members, seasons: seasons, badges: badges, tx: tx}
}

func (u *Prize) season(ctx context.Context) (*domain.Season, error) {
	s, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}
	return s, nil
}

func (u *Prize) List(ctx context.Context, status string) ([]domain.Prize, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}
	return u.prizes.List(ctx, season.ID, status)
}

type PrizeInput struct {
	NamaHadiah     string
	Deskripsi      *string
	NilaiEstimasi  *float64
	Alokasi        string
	KategoriTarget *string
}

// Donate is the member-facing path: the item enters the pool pending, and only
// an admin's approval makes it count — including for the PATRON badge.
func (u *Prize) Donate(ctx context.Context, in PrizeInput, donorMemberID string) (string, error) {
	in.NamaHadiah = strings.TrimSpace(in.NamaHadiah)
	if in.NamaHadiah == "" {
		return "", domain.Invalid("Nama hadiah wajib diisi.")
	}
	if !domain.ValidPrizeAlokasi(in.Alokasi) {
		return "", domain.Invalid("Alokasi hadiah tidak valid.")
	}

	season, err := u.season(ctx)
	if err != nil {
		return "", err
	}

	return u.prizes.Create(ctx, &domain.Prize{
		SeasonID:       season.ID,
		NamaHadiah:     in.NamaHadiah,
		Deskripsi:      in.Deskripsi,
		NilaiEstimasi:  in.NilaiEstimasi,
		DonaturID:      &donorMemberID,
		Alokasi:        in.Alokasi,
		KategoriTarget: in.KategoriTarget,
		Status:         "pending",
	})
}

// Seed is the committee path: an item the organisers provide, approved on
// arrival since there is nobody to vet.
func (u *Prize) Seed(ctx context.Context, in PrizeInput) (string, error) {
	in.NamaHadiah = strings.TrimSpace(in.NamaHadiah)
	if in.NamaHadiah == "" {
		return "", domain.Invalid("Nama hadiah wajib diisi.")
	}
	if !domain.ValidPrizeAlokasi(in.Alokasi) {
		return "", domain.Invalid("Alokasi hadiah tidak valid.")
	}

	season, err := u.season(ctx)
	if err != nil {
		return "", err
	}

	return u.prizes.Create(ctx, &domain.Prize{
		SeasonID:       season.ID,
		NamaHadiah:     in.NamaHadiah,
		Deskripsi:      in.Deskripsi,
		NilaiEstimasi:  in.NilaiEstimasi,
		Alokasi:        in.Alokasi,
		KategoriTarget: in.KategoriTarget,
		Status:         "approved",
	})
}

// SetStatus approves, rejects or awards a prize. Approving a donation is what
// earns the donor their PATRON badge.
func (u *Prize) SetStatus(ctx context.Context, id, status string, pemenangID *string) error {
	if !domain.ValidPrizeStatus(status) {
		return domain.Invalid("Status hadiah tidak valid.")
	}

	prize, err := u.prizes.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if prize == nil {
		return domain.NotFound("Hadiah tidak ditemukan.")
	}
	if status == "awarded" && pemenangID == nil && prize.PemenangID == nil {
		return domain.Invalid("Pemenang wajib diisi saat menyerahkan hadiah.")
	}

	if err := u.prizes.SetStatus(ctx, id, status, pemenangID); err != nil {
		return err
	}

	if prize.DonaturID != nil {
		u.badges.EvaluateQuietly(ctx, *prize.DonaturID, prize.SeasonID)
	}
	return nil
}

// Draw picks the winner of a lottery-allocated prize. It is the payoff of the
// whole ticket system, and it is one-way: once a prize has a winner this
// refuses rather than quietly re-rolling, because the first result has
// already been announced in the room.
func (u *Prize) Draw(ctx context.Context, prizeID string) (*domain.Prize, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	prize, err := u.prizes.FindByID(ctx, prizeID)
	if err != nil {
		return nil, err
	}
	if prize == nil || prize.SeasonID != season.ID {
		return nil, domain.NotFound("Hadiah tidak ditemukan.")
	}
	if prize.Alokasi != "undian" {
		return nil, domain.Invalid("Hadiah ini dialokasikan per kategori, bukan diundi.")
	}
	// The winner check comes first: a drawn prize is already "awarded", and
	// reporting that as "not approved yet" would send the admin looking for
	// the wrong problem.
	if prize.PemenangID != nil {
		return nil, domain.Conflict("Hadiah ini sudah diundi.")
	}
	if prize.Status != "approved" {
		return nil, domain.Invalid("Hadiah harus berstatus approved sebelum diundi.")
	}

	won, err := u.prizes.DrawWinner(ctx, season.ID, prizeID)
	if err != nil {
		return nil, err
	}
	if won == nil {
		// Either no tickets exist yet or another admin drew first. Re-reading
		// tells the two apart, and the difference matters: one is fixed by
		// issuing tickets, the other by refreshing the page.
		current, err := u.prizes.FindByID(ctx, prizeID)
		if err != nil {
			return nil, err
		}
		if current != nil && current.PemenangID != nil {
			return nil, domain.Conflict("Hadiah ini sudah diundi.")
		}
		return nil, domain.Invalid("Belum ada tiket undian. Terbitkan tiket lebih dulu.")
	}
	return won, nil
}

func (u *Prize) Delete(ctx context.Context, id string) error {
	return u.prizes.Delete(ctx, id)
}

// TicketSummary is one member's standing in the raffle.
type TicketSummary struct {
	MemberID string  `json:"member_id"`
	FullName string  `json:"full_name"`
	NamaTim  *string `json:"nama_tim"`
	Tickets  int     `json:"tickets"`
}

// IssueTickets recomputes every member's entitlement from their current season
// totals. Rewriting rather than appending keeps the pass re-runnable without
// inflating anyone's odds.
func (u *Prize) IssueTickets(ctx context.Context) ([]TicketSummary, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	var counts []domain.TicketCount
	err = u.tx.WithinTx(ctx, func(ctx context.Context) error {
		counts, err = u.prizes.RebuildTickets(ctx, season.ID)
		return err
	})
	if err != nil {
		return nil, err
	}

	out := make([]TicketSummary, 0, len(counts))
	for _, c := range counts {
		out = append(out, TicketSummary{
			MemberID: c.MemberID,
			FullName: c.FullName,
			NamaTim:  c.NamaTim,
			Tickets:  c.Tickets,
		})
	}
	return out, nil
}

func (u *Prize) TicketStandings(ctx context.Context) ([]TicketSummary, error) {
	season, err := u.season(ctx)
	if err != nil {
		return nil, err
	}

	counts, err := u.prizes.TicketCounts(ctx, season.ID)
	if err != nil {
		return nil, err
	}

	members, err := u.members.ListBySeason(ctx, season.ID)
	if err != nil {
		return nil, err
	}

	out := []TicketSummary{}
	for _, member := range members {
		if n := counts[member.ID]; n > 0 {
			out = append(out, TicketSummary{
				MemberID: member.ID,
				FullName: member.FullName,
				NamaTim:  member.NamaTim,
				Tickets:  n,
			})
		}
	}
	return out, nil
}
