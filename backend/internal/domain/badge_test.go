package domain

import (
	"slices"
	"testing"
)

func TestEarnedBadgesOnAnEmptySeason(t *testing.T) {
	if got := EarnedBadges(BadgeStats{}); len(got) != 0 {
		t.Errorf("a blank season earned %v, want nothing", got)
	}
}

func TestEarnedBadgesThresholds(t *testing.T) {
	cases := []struct {
		name  string
		stats BadgeStats
		want  string
		got   bool
	}{
		{"first verified TYFCB", BadgeStats{VerifiedTyfcbCount: 1}, BadgeFirstBlood, true},
		{"first visitor attending", BadgeStats{VisitorsHadir: 1}, BadgeHost, true},
		{"first conversion", BadgeStats{Conversions: 1}, BadgeCloser, true},

		{"four receivers is not a connector", BadgeStats{DistinctReceivers: 4}, BadgeConnector, false},
		{"five receivers is", BadgeStats{DistinctReceivers: 5}, BadgeConnector, true},
		{"nine receivers is not a spreader", BadgeStats{DistinctReceivers: 9}, BadgeSpreader, false},
		{"ten receivers is", BadgeStats{DistinctReceivers: 10}, BadgeSpreader, true},

		{"99 points is not a centurion", BadgeStats{ScoreOverall: 99}, BadgeCenturion, false},
		{"100 points is", BadgeStats{ScoreOverall: 100}, BadgeCenturion, true},

		{"two full attendances is not a hat-trick", BadgeStats{VisitorsHadirPenuh: 2}, BadgeHatTrick, false},
		{"three is", BadgeStats{VisitorsHadirPenuh: 3}, BadgeHatTrick, true},

		{"just under 250 juta", BadgeStats{LargestTyfcb: 249_999_999}, BadgeHighRoller, false},
		{"exactly 250 juta", BadgeStats{LargestTyfcb: 250_000_000}, BadgeHighRoller, true},

		{"two scoring days is not a streak", BadgeStats{DistinctScoringDays: 2}, BadgeStreakMaster, false},
		{"three is", BadgeStats{DistinctScoringDays: 3}, BadgeStreakMaster, true},

		{"colour raised", BadgeStats{ColorStatusRaised: true}, BadgeLevelUp, true},
		{"full roster contribution", BadgeStats{ContributedFullRoster: true}, BadgeTeamPlayer, true},
		{"approved prize", BadgeStats{ApprovedPrizes: 1}, BadgePatron, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			earned := slices.Contains(EarnedBadges(c.stats), c.want)
			if earned != c.got {
				t.Errorf("%s earned=%v, want %v", c.want, earned, c.got)
			}
		})
	}
}

// Spreader implies Connector, so a prolific member holds both rather than
// silently losing the lesser one.
func TestSpreaderAlsoEarnsConnector(t *testing.T) {
	earned := EarnedBadges(BadgeStats{DistinctReceivers: 12})

	for _, code := range []string{BadgeConnector, BadgeSpreader} {
		if !slices.Contains(earned, code) {
			t.Errorf("missing %s in %v", code, earned)
		}
	}
}

func TestEarnedBadgesIsPure(t *testing.T) {
	// Re-deriving must be stable: the evaluator calls this after every scoring
	// change and relies on idempotent awarding.
	stats := BadgeStats{VerifiedTyfcbCount: 3, DistinctReceivers: 6, ScoreOverall: 150}

	first := EarnedBadges(stats)
	second := EarnedBadges(stats)

	if !slices.Equal(first, second) {
		t.Errorf("not deterministic: %v then %v", first, second)
	}
}

func TestColorRank(t *testing.T) {
	if ColorRank(ColorMerah) >= ColorRank(ColorKuning) {
		t.Error("kuning must outrank merah")
	}
	if ColorRank(ColorKuning) >= ColorRank(ColorHijau) {
		t.Error("hijau must outrank kuning")
	}
	if ColorRank("unknown") != 0 {
		t.Error("an unknown status must rank below every real one")
	}
}
