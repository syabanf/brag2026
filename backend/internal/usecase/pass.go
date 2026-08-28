package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// ScoringPass settles the bonuses that depend on a whole day or week rather
// than a single submission. Every pass is keyed by period so re-running it is
// a no-op — the committee can trigger it as often as they like.
type ScoringPass struct {
	passes  domain.ScoringPassRepository
	ledger  domain.LedgerRepository
	events  domain.WeeklyEventRepository
	seasons domain.SeasonRepository
	badges  *Badges
	tx      domain.TxManager
}

func NewScoringPass(
	passes domain.ScoringPassRepository,
	ledger domain.LedgerRepository,
	events domain.WeeklyEventRepository,
	seasons domain.SeasonRepository,
	badges *Badges,
	tx domain.TxManager,
) *ScoringPass {
	return &ScoringPass{
		passes: passes, ledger: ledger, events: events,
		seasons: seasons, badges: badges, tx: tx,
	}
}

// PassResult reports what a run actually settled, so the admin screen can show
// the outcome rather than a bare "done".
type PassResult struct {
	Period       string   `json:"period"`
	FullRoster   []string `json:"full_roster_teams"`
	StreakAwards int      `json:"streak_awards"`
	HighRoller   *string  `json:"high_roller_member,omitempty"`
	PointsAdded  int      `json:"points_added"`
	Skipped      []string `json:"skipped,omitempty"`
}

// weekBounds returns the Monday–Sunday window containing day.
func weekBounds(day time.Time) (time.Time, time.Time) {
	offset := (int(day.Weekday()) + 6) % 7 // Monday = 0
	start := day.AddDate(0, 0, -offset).Truncate(24 * time.Hour)
	return start, start.AddDate(0, 0, 6)
}

// RunWeekly settles Full Roster for every qualifying team and, when a STREAK
// event covered the week, the flat streak bonus.
func (u *ScoringPass) RunWeekly(ctx context.Context, day time.Time) (*PassResult, error) {
	season, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if season == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}

	from, to := weekBounds(day)
	result := &PassResult{
		Period:     fmt.Sprintf("%s — %s", from.Format("2006-01-02"), to.Format("2006-01-02")),
		FullRoster: []string{},
	}

	rosters, err := u.passes.RosterForWeek(ctx, season.ID, from, to)
	if err != nil {
		return nil, err
	}

	// A STREAK event pays a flat bonus that the PRD explicitly excludes from
	// the event pengali.
	event, err := u.events.ActiveOn(ctx, season.ID, day)
	if err != nil {
		return nil, err
	}
	streakActive := event != nil && event.EventCode == domain.EventStreak

	var streakMembers []string
	if streakActive {
		streakMembers, err = u.passes.MembersWithScoringDays(ctx, season.ID, from, to, 3)
		if err != nil {
			return nil, err
		}
	}

	err = u.tx.WithinTx(ctx, func(ctx context.Context) error {
		for _, roster := range rosters {
			if !roster.QualifiesFullRoster() {
				continue
			}

			ref := fmt.Sprintf("full_roster:%s:%s", roster.TeamID, from.Format("2006-01-02"))
			settled, err := u.passes.AlreadySettled(ctx, ref)
			if err != nil {
				return err
			}
			if settled {
				result.Skipped = append(result.Skipped, roster.NamaTim+" (sudah dibayar)")
				continue
			}

			keterangan := fmt.Sprintf("Full Roster minggu %s", from.Format("2 Jan"))
			teamID := roster.TeamID

			if err := u.ledger.Append(ctx, &domain.LedgerEntry{
				SeasonID:   season.ID,
				TeamID:     &teamID,
				Kategori:   domain.CategoryBonus,
				Points:     domain.BonusFullRoster,
				SumberRef:  &ref,
				Keterangan: &keterangan,
			}); err != nil {
				return err
			}

			result.FullRoster = append(result.FullRoster, roster.NamaTim)
			result.PointsAdded += domain.BonusFullRoster
		}

		for _, memberID := range streakMembers {
			ref := fmt.Sprintf("streak:%s:%s", memberID, from.Format("2006-01-02"))
			settled, err := u.passes.AlreadySettled(ctx, ref)
			if err != nil {
				return err
			}
			if settled {
				continue
			}

			teamID, err := u.passes.TeamOf(ctx, memberID)
			if err != nil {
				return err
			}

			id := memberID
			keterangan := "Streak Week: 3+ hari aktif"

			if err := u.ledger.Append(ctx, &domain.LedgerEntry{
				SeasonID:   season.ID,
				MemberID:   &id,
				TeamID:     teamID,
				Kategori:   domain.CategoryBonus,
				Points:     domain.FlatStreak,
				SumberRef:  &ref,
				Keterangan: &keterangan,
			}); err != nil {
				return err
			}

			result.StreakAwards++
			result.PointsAdded += domain.FlatStreak
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Full Roster unlocks TEAM_PLAYER, so everyone on a qualifying team is
	// re-evaluated once the points have landed.
	for _, roster := range rosters {
		if roster.QualifiesFullRoster() {
			for _, memberID := range roster.MemberIDs {
				u.badges.EvaluateQuietly(ctx, memberID, season.ID)
			}
		}
	}

	return result, nil
}

// RunDaily settles the HIGH_ROLLER flat bonus for the day's single largest
// verified TYFCB, when that event is the one running.
func (u *ScoringPass) RunDaily(ctx context.Context, day time.Time) (*PassResult, error) {
	season, err := u.seasons.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if season == nil {
		return nil, domain.NotFound("Season aktif tidak ditemukan.")
	}

	result := &PassResult{Period: day.Format("2006-01-02"), FullRoster: []string{}}

	event, err := u.events.ActiveOn(ctx, season.ID, day)
	if err != nil {
		return nil, err
	}
	if event == nil || event.EventCode != domain.EventHighRoller {
		result.Skipped = append(result.Skipped, "High Roller Day tidak aktif hari ini")
		return result, nil
	}

	top, err := u.passes.TopTyfcbOfDay(ctx, season.ID, day)
	if err != nil {
		return nil, err
	}
	if top == nil {
		result.Skipped = append(result.Skipped, "tidak ada TYFCB verified hari ini")
		return result, nil
	}

	ref := fmt.Sprintf("high_roller:%s", day.Format("2006-01-02"))
	settled, err := u.passes.AlreadySettled(ctx, ref)
	if err != nil {
		return nil, err
	}
	if settled {
		result.Skipped = append(result.Skipped, "sudah dibayar")
		return result, nil
	}

	teamID, err := u.passes.TeamOf(ctx, top.GiverID)
	if err != nil {
		return nil, err
	}

	err = u.tx.WithinTx(ctx, func(ctx context.Context) error {
		keterangan := fmt.Sprintf("High Roller Day %s", day.Format("2 Jan"))
		giver := top.GiverID

		return u.ledger.Append(ctx, &domain.LedgerEntry{
			SeasonID:   season.ID,
			MemberID:   &giver,
			TeamID:     teamID,
			Kategori:   domain.CategoryBonus,
			Points:     domain.FlatHighRoller,
			SumberRef:  &ref,
			Keterangan: &keterangan,
		})
	})
	if err != nil {
		return nil, err
	}

	name := top.GiverName
	result.HighRoller = &name
	result.PointsAdded = domain.FlatHighRoller

	u.badges.EvaluateQuietly(ctx, top.GiverID, season.ID)
	return result, nil
}
