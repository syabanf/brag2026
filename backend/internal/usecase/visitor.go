package usecase

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type Visitor struct {
	visitors domain.VisitorRepository
	members  domain.MemberRepository
	ledger   domain.LedgerRepository
	events   domain.WeeklyEventRepository
	badges   *Badges
	tx       domain.TxManager
}

func NewVisitor(
	visitors domain.VisitorRepository,
	members domain.MemberRepository,
	ledger domain.LedgerRepository,
	events domain.WeeklyEventRepository,
	badges *Badges,
	tx domain.TxManager,
) *Visitor {
	return &Visitor{
		visitors: visitors, members: members, ledger: ledger,
		events: events, badges: badges, tx: tx,
	}
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

	// Milestones are credited when the admin confirms them, so the pengali is
	// the one running today rather than on the invitation date.
	event, err := u.events.ActiveOn(ctx, visitor.SeasonID, time.Now())
	if err != nil {
		return err
	}

	err = u.tx.WithinTx(ctx, func(ctx context.Context) error {
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

				delta := domain.VisitorStatusDelta(from, to)
				if delta > 0 {
					// Only awards are boosted; see the reversal branch below.
					delta = applyMultiplier(delta, domain.VisitorMultiplier(event, false))
				} else if delta < 0 {
					// Reverse to the new base rather than recomputing at today's
					// pengali, which may differ from the one that granted it.
					credited, err := u.ledger.SumBySource(ctx, visitor.ID)
					if err != nil {
						return err
					}
					delta = domain.VisitorCumulative(to) - credited
				}

				if delta != 0 {
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

			var points int
			keterangan := "Visitor konversi"
			if to {
				points = applyMultiplier(domain.ConversionPoints, domain.VisitorMultiplier(event, true))
			} else {
				// Give back exactly what the conversion was worth when granted.
				credited, err := u.ledger.SumBySource(ctx, conversionRef(visitor.ID))
				if err != nil {
					return err
				}
				points = -credited
				keterangan = "Pembatalan konversi visitor"
			}
			ref := conversionRef(visitor.ID)

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
	if err != nil {
		return err
	}

	u.badges.EvaluateQuietly(ctx, visitor.InviterID, visitor.SeasonID)
	return nil
}

// conversionRef keys the conversion bonus separately from attendance
// milestones, so correcting one never disturbs the other.
func conversionRef(visitorID string) string { return visitorID + ":conversion" }

// applyMultiplier scales a point delta and keeps the sign, so reversing a
// boosted milestone gives back exactly what was awarded.
func applyMultiplier(points int, multiplier float64) int {
	if multiplier == 1 {
		return points
	}
	return int(math.Round(float64(points) * multiplier))
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

	// What the visitor actually earned, boosts included — recomputing from the
	// status would under-refund anything credited during an event week.
	attendance, err := u.ledger.SumBySource(ctx, visitor.ID)
	if err != nil {
		return err
	}
	conversion, err := u.ledger.SumBySource(ctx, conversionRef(visitor.ID))
	if err != nil {
		return err
	}
	earned := attendance + conversion

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
