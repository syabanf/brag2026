package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type Visitor struct {
	visitors domain.VisitorRepository
	members  domain.MemberRepository
	ledger   domain.LedgerRepository
	tx       domain.TxManager
}

func NewVisitor(
	visitors domain.VisitorRepository,
	members domain.MemberRepository,
	ledger domain.LedgerRepository,
	tx domain.TxManager,
) *Visitor {
	return &Visitor{visitors: visitors, members: members, ledger: ledger, tx: tx}
}

type RegisterVisitorInput struct {
	Nama          string
	Kontak        string
	TanggalUndang time.Time
	InviterID     string
}

func (u *Visitor) Register(ctx context.Context, in RegisterVisitorInput, submittedBy *string) (*domain.Visitor, error) {
	in.Nama = strings.TrimSpace(in.Nama)
	in.Kontak = strings.TrimSpace(in.Kontak)

	if in.Nama == "" || in.Kontak == "" || in.TanggalUndang.IsZero() {
		return nil, domain.Invalid("nama, kontak, dan tanggal_undang wajib diisi.")
	}

	inviter, err := u.members.FindByID(ctx, in.InviterID)
	if err != nil {
		return nil, err
	}
	if inviter == nil {
		return nil, domain.NotFound("Profil member tidak ditemukan di season ini.")
	}

	visitor := &domain.Visitor{
		SeasonID:      inviter.SeasonID,
		Nama:          in.Nama,
		Kontak:        in.Kontak,
		InviterID:     inviter.ID,
		TanggalUndang: in.TanggalUndang,
		StatusHadir:   domain.VisitorTerdaftar,
	}

	id, err := u.visitors.Create(ctx, visitor, submittedBy)
	if err != nil {
		return nil, err
	}
	visitor.ID = id

	return visitor, nil
}

type UpdateVisitorInput struct {
	StatusHadir *string
	IsConverted *bool
}

// Update moves attendance status and/or the conversion flag. Points follow the
// difference between cumulative values, so a correction downwards reverses
// exactly what was awarded — no separate undo path.
func (u *Visitor) Update(ctx context.Context, id string, in UpdateVisitorInput) error {
	if in.StatusHadir == nil && in.IsConverted == nil {
		return domain.Invalid("Tidak ada perubahan.")
	}
	if in.StatusHadir != nil && !domain.ValidVisitorStatus(*in.StatusHadir) {
		return domain.Invalid("Status tidak valid.")
	}

	visitor, err := u.visitors.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if visitor == nil {
		return domain.NotFound("Visitor tidak ditemukan.")
	}
	if visitor.IsVoid {
		return domain.Conflict("Visitor sudah di-void dan tidak bisa diubah.")
	}

	inviter, err := u.members.FindByID(ctx, visitor.InviterID)
	if err != nil {
		return err
	}
	var teamID *string
	if inviter != nil {
		teamID = inviter.TeamID
	}

	return u.tx.WithinTx(ctx, func(ctx context.Context) error {
		if in.StatusHadir != nil {
			to := domain.VisitorStatus(*in.StatusHadir)
			from := visitor.StatusHadir

			if to != from {
				// The guarded update is optimistic locking: a concurrent request
				// that already moved the status loses, so points never double.
				ok, err := u.visitors.UpdateStatusGuarded(ctx, id, from, to)
				if err != nil {
					return err
				}
				if !ok {
					return domain.Conflict("Status visitor sudah berubah. Muat ulang halaman.")
				}

				if delta := domain.VisitorStatusDelta(from, to); delta != 0 {
					ref := visitor.ID
					keterangan := fmt.Sprintf("Status visitor: %s → %s",
						domain.VisitorStatusLabel(from), domain.VisitorStatusLabel(to))

					if err := u.ledger.Append(ctx, &domain.LedgerEntry{
						SeasonID:   visitor.SeasonID,
						MemberID:   &visitor.InviterID,
						TeamID:     teamID,
						Kategori:   domain.CategoryVisitor,
						Points:     delta,
						SumberRef:  &ref,
						Keterangan: &keterangan,
					}); err != nil {
						return err
					}
				}
			}
		}

		if in.IsConverted != nil && *in.IsConverted != visitor.IsConverted {
			to := *in.IsConverted

			ok, err := u.visitors.UpdateConversionGuarded(ctx, id, visitor.IsConverted, to)
			if err != nil {
				return err
			}
			if !ok {
				return domain.Conflict("Status konversi sudah berubah. Muat ulang halaman.")
			}

			points := domain.ConversionPoints
			keterangan := "Visitor konversi"
			if !to {
				points = -points
				keterangan = "Pembatalan konversi visitor"
			}
			ref := visitor.ID

			if err := u.ledger.Append(ctx, &domain.LedgerEntry{
				SeasonID:   visitor.SeasonID,
				MemberID:   &visitor.InviterID,
				TeamID:     teamID,
				Kategori:   domain.CategoryVisitor,
				Points:     points,
				SumberRef:  &ref,
				Keterangan: &keterangan,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

// Void cancels a visitor and reverses every point it has earned so far,
// including the conversion bonus.
func (u *Visitor) Void(ctx context.Context, id, actorID string) error {
	visitor, err := u.visitors.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if visitor == nil {
		return domain.NotFound("Visitor tidak ditemukan.")
	}
	if visitor.IsVoid {
		return domain.Conflict("Visitor sudah di-void.")
	}

	inviter, err := u.members.FindByID(ctx, visitor.InviterID)
	if err != nil {
		return err
	}
	var teamID *string
	if inviter != nil {
		teamID = inviter.TeamID
	}

	earned := domain.VisitorCumulative(visitor.StatusHadir)
	if visitor.IsConverted {
		earned += domain.ConversionPoints
	}

	return u.tx.WithinTx(ctx, func(ctx context.Context) error {
		if earned != 0 {
			ref := visitor.ID
			keterangan := "Visitor void"

			if err := u.ledger.Append(ctx, &domain.LedgerEntry{
				SeasonID:   visitor.SeasonID,
				MemberID:   &visitor.InviterID,
				TeamID:     teamID,
				Kategori:   domain.CategoryVisitor,
				Points:     -earned,
				SumberRef:  &ref,
				Keterangan: &keterangan,
			}); err != nil {
				return err
			}
		}
		return u.visitors.Void(ctx, id, actorID)
	})
}

func (u *Visitor) List(ctx context.Context, f domain.VisitorFilter) ([]domain.Visitor, error) {
	return u.visitors.List(ctx, f)
}
