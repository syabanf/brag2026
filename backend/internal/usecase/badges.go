package usecase

import (
	"context"
	"log/slog"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// Badges awards milestone badges. It is deliberately separate from the
// scoring use cases: a badge is a consequence of a score changing, never a
// precondition for one.
type Badges struct {
	badges domain.BadgeRepository
}

func NewBadges(badges domain.BadgeRepository) *Badges {
	return &Badges{badges: badges}
}

// Evaluate re-derives every badge a member qualifies for and awards what is
// missing. Awarding is idempotent, so this is safe to call after any scoring
// event without tracking what was already given.
func (u *Badges) Evaluate(ctx context.Context, memberID, seasonID string) error {
	stats, err := u.badges.Stats(ctx, memberID, seasonID)
	if err != nil {
		return err
	}

	for _, code := range domain.EarnedBadges(stats) {
		if err := u.badges.Award(ctx, memberID, code); err != nil {
			return err
		}
	}

	return nil
}

// EvaluateQuietly runs Evaluate but only logs a failure. Callers use it after
// a committed score change, where losing a badge is a far smaller problem than
// failing the request that earned it.
func (u *Badges) EvaluateQuietly(ctx context.Context, memberID, seasonID string) {
	if err := u.Evaluate(ctx, memberID, seasonID); err != nil {
		slog.Error("badge evaluation failed", "member_id", memberID, "err", err)
	}
}
