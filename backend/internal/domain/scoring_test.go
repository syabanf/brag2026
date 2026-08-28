package domain

import "testing"

func TestBand(t *testing.T) {
	// Boundaries are exclusive at the top, so an exact threshold lands in the
	// band above it.
	cases := []struct {
		nilai float64
		want  int
	}{
		{0, 10},
		{499_999, 10},
		{500_000, 25},
		{1_999_999, 25},
		{2_000_000, 50},
		{9_999_999, 50},
		{10_000_000, 80},
		{49_999_999, 80},
		{50_000_000, 120},
		{249_999_999, 120},
		{250_000_000, 150},
		{499_999_999, 150},
		{500_000_000, 200},
		{1_000_000_000, 200},
	}

	for _, c := range cases {
		if got := Band(c.nilai); got != c.want {
			t.Errorf("Band(%.0f) = %d, want %d", c.nilai, got, c.want)
		}
	}
}

func TestPairPenalty(t *testing.T) {
	cases := []struct {
		ordinal int
		want    float64
	}{
		{1, 1},
		{2, 0.5},
		{4, 0.25},
		{0, 1},  // guarded: an unset ordinal must not divide by zero
		{-3, 1}, // guarded likewise
	}

	for _, c := range cases {
		if got := PairPenalty(c.ordinal); got != c.want {
			t.Errorf("PairPenalty(%d) = %v, want %v", c.ordinal, got, c.want)
		}
	}
}

func TestTyfcbScore(t *testing.T) {
	cases := []struct {
		name       string
		nilai      float64
		ordinal    int
		multiplier float64
		want       int
	}{
		{"first transaction at full band", 18_000_000, 1, 1, 80},
		{"second of a pair is halved", 18_000_000, 2, 1, 40},
		{"third of a pair rounds up", 18_000_000, 3, 1, 27},
		{"double event doubles the score", 18_000_000, 1, 2, 160},
		{"penalty and multiplier compose", 18_000_000, 2, 1.5, 60},
		{"top band", 780_000_000, 1, 1, 200},
		{"smallest band", 350_000, 1, 1, 10},
		{"zero multiplier falls back to 1", 18_000_000, 1, 0, 80},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := TyfcbScore(c.nilai, c.ordinal, c.multiplier); got != c.want {
				t.Errorf("TyfcbScore(%.0f, %d, %v) = %d, want %d",
					c.nilai, c.ordinal, c.multiplier, got, c.want)
			}
		})
	}
}

func TestVisitorStatusDelta(t *testing.T) {
	cases := []struct {
		from, to VisitorStatus
		want     int
	}{
		{VisitorTerdaftar, VisitorHadir, 20},
		{VisitorHadir, VisitorHadirPenuh, 30},
		{VisitorTerdaftar, VisitorHadirPenuh, 50},
		// Corrections downwards must reverse exactly what was awarded.
		{VisitorHadirPenuh, VisitorHadir, -30},
		{VisitorHadir, VisitorTerdaftar, -20},
		{VisitorHadirPenuh, VisitorTerdaftar, -50},
		{VisitorHadir, VisitorHadir, 0},
	}

	for _, c := range cases {
		if got := VisitorStatusDelta(c.from, c.to); got != c.want {
			t.Errorf("VisitorStatusDelta(%s, %s) = %d, want %d", c.from, c.to, got, c.want)
		}
	}
}

// A round trip up and back down must net to zero, or a leaderboard would drift
// every time an admin corrected a mistake.
func TestVisitorStatusRoundTripIsNeutral(t *testing.T) {
	path := []VisitorStatus{
		VisitorTerdaftar, VisitorHadir, VisitorHadirPenuh, VisitorHadir, VisitorTerdaftar,
	}

	total := 0
	for i := 1; i < len(path); i++ {
		total += VisitorStatusDelta(path[i-1], path[i])
	}

	if total != 0 {
		t.Errorf("round trip netted %d points, want 0", total)
	}
}

func TestVisitorCumulative(t *testing.T) {
	if got := VisitorCumulative(VisitorHadirPenuh); got != 50 {
		t.Errorf("VisitorCumulative(hadir_penuh) = %d, want 50", got)
	}
	if got := VisitorCumulative("unknown"); got != 0 {
		t.Errorf("VisitorCumulative(unknown) = %d, want 0", got)
	}
}

func TestRoleHierarchy(t *testing.T) {
	if !RoleAdmin.IsAdmin() {
		t.Error("admin should be admin")
	}
	if RoleCaptain.IsAdmin() {
		t.Error("captain must not be admin")
	}
	// An admin can do everything a captain can.
	if !RoleAdmin.IsCaptain() {
		t.Error("admin should satisfy captain checks")
	}
	if RoleMember.IsCaptain() {
		t.Error("member must not satisfy captain checks")
	}
}
