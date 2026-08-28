package domain

import (
	"testing"
	"time"
)

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func event(code EventCode) *WeeklyEvent {
	return &WeeklyEvent{
		EventCode:      code,
		TanggalMulai:   day("2026-09-14"),
		TanggalSelesai: day("2026-09-20"),
	}
}

func TestTyfcbMultiplierWithoutEvent(t *testing.T) {
	// No event scheduled must leave scoring untouched.
	if got := TyfcbMultiplier(nil, TyfcbContext{Tanggal: day("2026-09-15")}); got != 1 {
		t.Errorf("no event = %v, want 1", got)
	}
}

func TestTyfcbMultiplier(t *testing.T) {
	klasA, klasB := "klas-a", "klas-b"

	carousel := event(EventCatCarousel)
	carousel.TargetClassificationID = &klasA

	cases := []struct {
		name  string
		event *WeeklyEvent
		ctx   TyfcbContext
		want  float64
	}{
		{"carousel hits the target classification", carousel,
			TyfcbContext{ReceiverClassificationID: &klasA}, 2},
		{"carousel ignores another classification", carousel,
			TyfcbContext{ReceiverClassificationID: &klasB}, 1},
		{"carousel ignores an unclassified member", carousel,
			TyfcbContext{}, 1},

		{"spread love rewards a new pair", event(EventSpreadLove),
			TyfcbContext{PairOrdinal: 1}, 2},
		{"spread love ignores a repeat pair", event(EventSpreadLove),
			TyfcbContext{PairOrdinal: 2}, 1},

		{"underdog rewards merah", event(EventUnderdog),
			TyfcbContext{ReceiverColorStatus: ColorMerah}, 2},
		{"underdog rewards kuning", event(EventUnderdog),
			TyfcbContext{ReceiverColorStatus: ColorKuning}, 2},
		{"underdog ignores hijau", event(EventUnderdog),
			TyfcbContext{ReceiverColorStatus: ColorHijau}, 1},

		{"double-up on saturday", event(EventDoubleUp),
			TyfcbContext{Tanggal: day("2026-09-19")}, 1.5},
		{"double-up on sunday", event(EventDoubleUp),
			TyfcbContext{Tanggal: day("2026-09-20")}, 1.5},
		{"double-up ignores a weekday", event(EventDoubleUp),
			TyfcbContext{Tanggal: day("2026-09-16")}, 1},

		{"founder boosts everything", event(EventFounder), TyfcbContext{}, 1.5},

		{"power team rewards a shared sphere", event(EventPowerTeam),
			TyfcbContext{SameContactSphere: true}, 1.5},
		{"power team ignores unrelated trades", event(EventPowerTeam),
			TyfcbContext{SameContactSphere: false}, 1},

		// Flat-bonus events are settled by a periodic pass, so the PRD keeps
		// them out of the multiplier entirely.
		{"one-to-one stays neutral", event(EventOneToOne), TyfcbContext{}, 1},
		{"high roller stays neutral", event(EventHighRoller), TyfcbContext{}, 1},
		{"streak stays neutral", event(EventStreak), TyfcbContext{}, 1},
		// A sphere overlap must not boost weeks that are not POWER_TEAM.
		{"sphere overlap is ignored outside power team", event(EventUnderdog),
			TyfcbContext{SameContactSphere: true, ReceiverColorStatus: ColorHijau}, 1},
		// Visitor-side events must not touch TYFCB.
		{"visitor blitz stays neutral", event(EventVisitorBlitz), TyfcbContext{}, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := TyfcbMultiplier(c.event, c.ctx); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestTyfcbMultiplierComposesWithPairPenalty(t *testing.T) {
	// The three factors must multiply, not override one another:
	// band 80 × pair 1/2 × event 1.5 = 60.
	if got := TyfcbScore(18_000_000, 2, 1.5); got != 60 {
		t.Errorf("composed score = %d, want 60", got)
	}
}

func TestVisitorMultiplier(t *testing.T) {
	cases := []struct {
		name       string
		event      *WeeklyEvent
		conversion bool
		want       float64
	}{
		{"no event", nil, false, 1},
		{"blitz boosts attendance", event(EventVisitorBlitz), false, 1.5},
		{"blitz boosts conversion too", event(EventVisitorBlitz), true, 1.5},
		{"new blood boosts attendance only", event(EventNewBlood), false, 2},
		{"new blood leaves conversion alone", event(EventNewBlood), true, 1},
		{"closing week boosts conversion only", event(EventClosingWeek), true, 2},
		{"closing week leaves attendance alone", event(EventClosingWeek), false, 1},
		{"founder boosts both", event(EventFounder), true, 1.5},
		// TYFCB-side events must not touch visitor points.
		{"underdog stays neutral", event(EventUnderdog), false, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := VisitorMultiplier(c.event, c.conversion); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestEventCovers(t *testing.T) {
	e := event(EventFounder)

	cases := []struct {
		day  string
		want bool
	}{
		{"2026-09-13", false}, // day before
		{"2026-09-14", true},  // first day, inclusive
		{"2026-09-17", true},
		{"2026-09-20", true}, // last day, inclusive
		{"2026-09-21", false},
	}

	for _, c := range cases {
		if got := e.Covers(day(c.day)); got != c.want {
			t.Errorf("Covers(%s) = %v, want %v", c.day, got, c.want)
		}
	}

	if (*WeeklyEvent)(nil).Covers(day("2026-09-15")) {
		t.Error("a nil event must cover nothing")
	}
}

func TestEventCatalogIsComplete(t *testing.T) {
	// Every code in the bank needs copy, or the UI would show a bare code.
	codes := EventCodes()
	if len(codes) != 12 {
		t.Fatalf("event bank has %d codes, want 12", len(codes))
	}

	for _, code := range codes {
		nama, mekanik := DescribeEvent(code)
		if nama == string(code) || mekanik == "" {
			t.Errorf("%s has no description", code)
		}
		if !ValidEventCode(string(code)) {
			t.Errorf("%s is not accepted as valid", code)
		}
	}

	if ValidEventCode("NOT_AN_EVENT") {
		t.Error("unknown codes must be rejected")
	}
}
