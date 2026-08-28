package domain

import "testing"

func TestLevelUpBonus(t *testing.T) {
	cases := []struct {
		from, to ColorStatus
		want     int
	}{
		{ColorMerah, ColorKuning, 75},
		{ColorKuning, ColorHijau, 150},
		// Two steps at once still pays both.
		{ColorMerah, ColorHijau, 225},

		// Corrections downwards pay nothing rather than clawing back: the team
		// never chose the demotion.
		{ColorHijau, ColorKuning, 0},
		{ColorKuning, ColorMerah, 0},
		{ColorHijau, ColorMerah, 0},

		// No movement, no bonus.
		{ColorMerah, ColorMerah, 0},
		{ColorHijau, ColorHijau, 0},
	}

	for _, c := range cases {
		if got := LevelUpBonus(c.from, c.to); got != c.want {
			t.Errorf("LevelUpBonus(%s, %s) = %d, want %d", c.from, c.to, got, c.want)
		}
	}
}

func TestQualifiesFullRoster(t *testing.T) {
	cases := []struct {
		name   string
		roster RosterStatus
		want   bool
	}{
		{"everyone scored", RosterStatus{ActiveCount: 5, ScoringCount: 5}, true},
		{"one short", RosterStatus{ActiveCount: 5, ScoringCount: 4}, false},
		{"nobody scored", RosterStatus{ActiveCount: 5, ScoringCount: 0}, false},
		{"single-member team", RosterStatus{ActiveCount: 1, ScoringCount: 1}, true},
		// An empty team must not collect the bonus every week for doing nothing.
		{"no active members", RosterStatus{ActiveCount: 0, ScoringCount: 0}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.roster.QualifiesFullRoster(); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestRaffleTickets(t *testing.T) {
	cases := []struct {
		name                            string
		score, visitors, newPairs, want int
	}{
		{"nothing earned", 0, 0, 0, 0},
		{"99 points is not a ticket yet", 99, 0, 0, 0},
		{"100 points is one", 100, 0, 0, 1},
		{"250 points floors to two", 250, 0, 0, 2},
		{"one per attending visitor", 0, 3, 0, 3},
		{"one per new pair", 0, 0, 4, 4},
		{"the three sources add up", 220, 2, 1, 5},
		// Defensive: a negative balance must not hand out negative tickets.
		{"negative score is clamped", -50, 1, 0, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := RaffleTickets(c.score, c.visitors, c.newPairs)
			if got != c.want {
				t.Errorf("RaffleTickets(%d, %d, %d) = %d, want %d",
					c.score, c.visitors, c.newPairs, got, c.want)
			}
		})
	}
}
