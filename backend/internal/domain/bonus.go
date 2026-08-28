package domain

// Team bonuses and flat bonuses are settled by a periodic pass rather than at
// submission time, because they depend on a whole week or day of activity.
// The PRD calls these out as NOT multiplied by the event pengali.
const (
	FlatHighRoller = 50
	FlatOneToOne   = 30
	FlatStreak     = 40
)

// LevelUpBonus is the team's reward when a member's colour status is raised.
// Only an upward step pays; a correction downwards is worth nothing rather
// than negative, since the team never chose it.
func LevelUpBonus(from, to ColorStatus) int {
	switch {
	case from == ColorMerah && to == ColorKuning:
		return 75
	case from == ColorKuning && to == ColorHijau:
		return 150
	case from == ColorMerah && to == ColorHijau:
		// Two steps at once still pays both.
		return 225
	}
	return 0
}

// RosterStatus is one team's Full Roster check for a week.
type RosterStatus struct {
	TeamID       string
	NamaTim      string
	ActiveCount  int
	ScoringCount int
	MemberIDs    []string
}

// QualifiesFullRoster is true when every active member scored at least once.
// A team with no active members does not qualify — otherwise an empty team
// would collect the bonus every week.
func (r RosterStatus) QualifiesFullRoster() bool {
	return r.ActiveCount > 0 && r.ScoringCount == r.ActiveCount
}

// RaffleTickets is the entitlement from the PRD: one per full 100 points, one
// per visitor who attended, and one per first-time pair.
func RaffleTickets(scoreOverall, visitorsAttended, newPairs int) int {
	if scoreOverall < 0 {
		scoreOverall = 0
	}
	return scoreOverall/100 + visitorsAttended + newPairs
}
