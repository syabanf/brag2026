package domain

import "time"

// EventCode identifies one of the twelve weekly events from the PRD's event
// bank. At most one is active per week per season.
type EventCode string

const (
	EventCatCarousel  EventCode = "CAT_CAROUSEL"
	EventVisitorBlitz EventCode = "VISITOR_BLITZ"
	EventClosingWeek  EventCode = "CLOSING_WEEK"
	EventSpreadLove   EventCode = "SPREAD_LOVE"
	EventUnderdog     EventCode = "UNDERDOG"
	EventPowerTeam    EventCode = "POWER_TEAM"
	EventHighRoller   EventCode = "HIGH_ROLLER"
	EventNewBlood     EventCode = "NEW_BLOOD"
	EventOneToOne     EventCode = "ONE_TO_ONE"
	EventDoubleUp     EventCode = "DOUBLE_UP"
	EventStreak       EventCode = "STREAK"
	EventFounder      EventCode = "FOUNDER"
)

// WeeklyEvent is the event the committee scheduled for one week.
type WeeklyEvent struct {
	ID                     string    `json:"id"`
	SeasonID               string    `json:"season_id"`
	MingguKe               int       `json:"minggu_ke"`
	EventCode              EventCode `json:"event_code"`
	TargetClassificationID *string   `json:"target_classification_id"`
	TanggalMulai           time.Time `json:"tanggal_mulai"`
	TanggalSelesai         time.Time `json:"tanggal_selesai"`
	Nama                   string    `json:"nama"`
	Mekanik                string    `json:"mekanik"`
}

// eventCatalog is the PRD's event bank, kept here so the API can describe an
// event without the frontend hardcoding Indonesian copy.
var eventCatalog = map[EventCode]struct{ Nama, Mekanik string }{
	EventCatCarousel:  {"Category Carousel", "TYFCB ke klasifikasi terpilih = 2×"},
	EventVisitorBlitz: {"Visitor Blitz", "Score visitor = 1.5×"},
	EventClosingWeek:  {"Closing Week", "Konversi member = 2×"},
	EventSpreadLove:   {"Spread the Love", "TYFCB ke receiver baru = 2×"},
	EventUnderdog:     {"Underdog Week", "TYFCB ke member merah/kuning = 2×"},
	EventPowerTeam:    {"Power Team Week", "TYFCB dalam contact sphere = 1.5×"},
	EventHighRoller:   {"High Roller Day", "TYFCB tunggal tertinggi hari itu = flat +50"},
	EventNewBlood:     {"New Blood", "Visitor milestone hadir = 2×"},
	EventOneToOne:     {"1-2-1 Payoff", "1-2-1 berujung TYFCB minggu sama = flat +30"},
	EventDoubleUp:     {"Double-Up Weekend", "Score Sabtu–Minggu = 1.5×"},
	EventStreak:       {"Streak Week", "Log 3+ hari berbeda = flat +40"},
	EventFounder:      {"Founder's Frenzy", "Semua score minggu itu = 1.5×"},
}

func DescribeEvent(code EventCode) (nama, mekanik string) {
	if meta, ok := eventCatalog[code]; ok {
		return meta.Nama, meta.Mekanik
	}
	return string(code), ""
}

func ValidEventCode(value string) bool {
	_, ok := eventCatalog[EventCode(value)]
	return ok
}

// EventCodes lists the bank in PRD order, for admin pickers.
func EventCodes() []EventCode {
	return []EventCode{
		EventCatCarousel, EventVisitorBlitz, EventClosingWeek, EventSpreadLove,
		EventUnderdog, EventPowerTeam, EventHighRoller, EventNewBlood,
		EventOneToOne, EventDoubleUp, EventStreak, EventFounder,
	}
}

// TyfcbContext carries the facts an event rule needs to judge one transaction.
// Passing a struct rather than the entity keeps the rule independent of how
// the row is stored.
type TyfcbContext struct {
	Tanggal time.Time
	// PairOrdinal is 1 for a pair's first-ever transaction.
	PairOrdinal int
	// ReceiverClassificationID is the classification of the member the
	// business went to.
	ReceiverClassificationID *string
	ReceiverColorStatus      ColorStatus
	// SameContactSphere is true when giver and receiver sit in a shared
	// contact sphere — complementary trades that naturally refer to each
	// other. Resolved by the repository, since it is a lookup rather than a
	// property of the transaction.
	SameContactSphere bool
}

// TyfcbMultiplier is the event pengali (M) for one transaction. The flat-bonus
// events (HIGH_ROLLER, ONE_TO_ONE, STREAK) are settled by a periodic pass
// instead, and the PRD excludes them from the multiplier by design, so they
// return 1 here.
func TyfcbMultiplier(event *WeeklyEvent, ctx TyfcbContext) float64 {
	if event == nil {
		return 1
	}

	switch event.EventCode {
	case EventCatCarousel:
		if event.TargetClassificationID != nil &&
			ctx.ReceiverClassificationID != nil &&
			*event.TargetClassificationID == *ctx.ReceiverClassificationID {
			return 2
		}
	case EventSpreadLove:
		if ctx.PairOrdinal == 1 {
			return 2
		}
	case EventUnderdog:
		if ctx.ReceiverColorStatus == ColorMerah || ctx.ReceiverColorStatus == ColorKuning {
			return 2
		}
	case EventDoubleUp:
		switch ctx.Tanggal.Weekday() {
		case time.Saturday, time.Sunday:
			return 1.5
		}
	case EventPowerTeam:
		if ctx.SameContactSphere {
			return 1.5
		}
	case EventFounder:
		return 1.5
	}

	return 1
}

// VisitorMultiplier is the event pengali for visitor points. `conversion`
// separates the +100 conversion bonus from ordinary attendance milestones,
// because CLOSING_WEEK doubles only the former.
func VisitorMultiplier(event *WeeklyEvent, conversion bool) float64 {
	if event == nil {
		return 1
	}

	switch event.EventCode {
	case EventVisitorBlitz:
		return 1.5
	case EventClosingWeek:
		if conversion {
			return 2
		}
	case EventNewBlood:
		if !conversion {
			return 2
		}
	case EventFounder:
		return 1.5
	}

	return 1
}

// Covers reports whether the event is running on the given date.
func (e *WeeklyEvent) Covers(day time.Time) bool {
	if e == nil {
		return false
	}
	d := day.Truncate(24 * time.Hour)
	return !d.Before(e.TanggalMulai.Truncate(24*time.Hour)) &&
		!d.After(e.TanggalSelesai.Truncate(24*time.Hour))
}
