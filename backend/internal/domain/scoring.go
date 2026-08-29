package domain

import "math"

// Band is the base point value for a TYFCB transaction, stepped by the rupiah
// amount closed. Boundaries are exclusive at the top: exactly 500_000 lands in
// the 25-point band.
func Band(nilai float64) int {
	switch {
	case nilai < 500_000:
		return 10
	case nilai < 2_000_000:
		return 25
	case nilai < 10_000_000:
		return 50
	case nilai < 50_000_000:
		return 80
	case nilai < 250_000_000:
		return 120
	case nilai < 500_000_000:
		return 150
	default:
		return 200
	}
}

// PairPenalty damps repeat business between the same two members: the nth
// transaction of a pair is worth 1/n of the band.
func PairPenalty(pairOrdinal int) float64 {
	if pairOrdinal < 1 {
		pairOrdinal = 1
	}
	return 1 / float64(pairOrdinal)
}

// TyfcbScore is the spec formula: round(Band × PairPenalty × EventMultiplier).
func TyfcbScore(nilai float64, pairOrdinal int, eventMultiplier float64) int {
	if eventMultiplier <= 0 {
		eventMultiplier = 1
	}
	return int(math.Round(float64(Band(nilai)) * PairPenalty(pairOrdinal) * eventMultiplier))
}

// visitorCumulative is what a visitor is worth in total once it reaches a
// status — not an increment. A status change awards the difference, which
// makes corrections in either direction fall out for free.
var visitorCumulative = map[VisitorStatus]int{
	VisitorTerdaftar:  0,
	VisitorHadir:      20,
	VisitorHadirPenuh: 50,
}

// ConversionPoints is the bonus for a visitor who becomes a member.
const ConversionPoints = 100

func VisitorCumulative(status VisitorStatus) int {
	return visitorCumulative[status]
}

// VisitorStatusDelta is the point movement for a status transition. It is
// negative when the status is corrected downwards.
func VisitorStatusDelta(from, to VisitorStatus) int {
	return visitorCumulative[to] - visitorCumulative[from]
}

var visitorStatusLabel = map[VisitorStatus]string{
	VisitorTerdaftar:  "Terdaftar",
	VisitorHadir:      "Hadir",
	VisitorHadirPenuh: "Hadir Penuh",
}

func VisitorStatusLabel(status VisitorStatus) string {
	if label, ok := visitorStatusLabel[status]; ok {
		return label
	}
	return string(status)
}

// Team bonuses, applied by the committee rather than earned per transaction.
const (
	BonusFullRoster   = 100
	BonusLevelUpSmall = 75
	BonusLevelUpLarge = 150
)

// ScoreCategory selects which of a season's running totals to rank by. The
// spec asks for six leaderboards: these three, each for teams and members.
type ScoreCategory string

const (
	ScoreOverall ScoreCategory = "overall"
	ScoreTyfcb   ScoreCategory = "tyfcb"
	ScoreVisitor ScoreCategory = "visitor"
)

// ParseScoreCategory falls back to overall, since an unrecognised tab in a URL
// should show the main board rather than an error page.
func ParseScoreCategory(v string) ScoreCategory {
	switch ScoreCategory(v) {
	case ScoreTyfcb:
		return ScoreTyfcb
	case ScoreVisitor:
		return ScoreVisitor
	default:
		return ScoreOverall
	}
}

// Column is the ledger total this category ranks by. It returns a fixed
// identifier from this package, never anything derived from a request, so it
// is safe to interpolate into a query.
func (c ScoreCategory) Column() string {
	switch c {
	case ScoreTyfcb:
		return "score_tyfcb"
	case ScoreVisitor:
		return "score_visitor"
	default:
		return "score_overall"
	}
}
