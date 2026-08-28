package domain

// Badge codes from the seed. Kept as constants so a typo is a compile error
// rather than a badge that silently never fires.
const (
	BadgeFirstBlood   = "FIRST_BLOOD"
	BadgeHost         = "HOST"
	BadgeCloser       = "CLOSER"
	BadgeConnector    = "CONNECTOR"
	BadgeSpreader     = "SPREADER"
	BadgeCenturion    = "CENTURION"
	BadgeHatTrick     = "HAT_TRICK"
	BadgeHighRoller   = "HIGH_ROLLER"
	BadgeStreakMaster = "STREAK_MASTER"
	BadgeTeamPlayer   = "TEAM_PLAYER"
	BadgeLevelUp      = "LEVEL_UP"
	BadgePatron       = "PATRON"
)

// BadgeStats is a snapshot of one member's season, gathered once so the rules
// below stay pure and instantly testable.
type BadgeStats struct {
	VerifiedTyfcbCount    int
	DistinctReceivers     int
	LargestTyfcb          float64
	ScoreOverall          int
	VisitorsHadir         int
	VisitorsHadirPenuh    int
	Conversions           int
	ColorStatusRaised     bool
	DistinctScoringDays   int
	ContributedFullRoster bool
	ApprovedPrizes        int
}

// Thresholds from the badge seed, named so the numbers are not scattered.
const (
	connectorReceivers = 5
	spreaderReceivers  = 10
	centurionScore     = 100
	hatTrickVisitors   = 3
	highRollerNilai    = 250_000_000
	streakDistinctDays = 3
)

// EarnedBadges returns every badge the stats qualify for. Awarding is
// idempotent upstream, so returning an already-held badge is harmless — that
// keeps this function a pure predicate over the season rather than a diff.
func EarnedBadges(s BadgeStats) []string {
	var earned []string

	add := func(ok bool, code string) {
		if ok {
			earned = append(earned, code)
		}
	}

	add(s.VerifiedTyfcbCount >= 1, BadgeFirstBlood)
	add(s.VisitorsHadir >= 1, BadgeHost)
	add(s.Conversions >= 1, BadgeCloser)
	add(s.DistinctReceivers >= connectorReceivers, BadgeConnector)
	add(s.DistinctReceivers >= spreaderReceivers, BadgeSpreader)
	add(s.ScoreOverall >= centurionScore, BadgeCenturion)
	add(s.VisitorsHadirPenuh >= hatTrickVisitors, BadgeHatTrick)
	add(s.LargestTyfcb >= highRollerNilai, BadgeHighRoller)
	add(s.DistinctScoringDays >= streakDistinctDays, BadgeStreakMaster)
	add(s.ContributedFullRoster, BadgeTeamPlayer)
	add(s.ColorStatusRaised, BadgeLevelUp)
	add(s.ApprovedPrizes >= 1, BadgePatron)

	return earned
}

// ColorRank orders the status ladder so a level-up can be detected.
func ColorRank(c ColorStatus) int {
	switch c {
	case ColorMerah:
		return 1
	case ColorKuning:
		return 2
	case ColorHijau:
		return 3
	}
	return 0
}
