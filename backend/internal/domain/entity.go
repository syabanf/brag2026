// Package domain holds the enterprise rules: entities, value objects and the
// repository contracts they need. It imports nothing from the outer layers.
package domain

import "time"

type Role string

const (
	RoleMember  Role = "member"
	RoleCaptain Role = "captain"
	RoleAdmin   Role = "admin"
)

func (r Role) IsAdmin() bool   { return r == RoleAdmin }
func (r Role) IsCaptain() bool { return r == RoleCaptain || r == RoleAdmin }

type ColorStatus string

const (
	ColorMerah  ColorStatus = "merah"
	ColorKuning ColorStatus = "kuning"
	ColorHijau  ColorStatus = "hijau"
)

func ValidColorStatus(v string) bool {
	switch ColorStatus(v) {
	case ColorMerah, ColorKuning, ColorHijau:
		return true
	}
	return false
}

type TyfcbStatus string

const (
	TyfcbPending  TyfcbStatus = "pending"
	TyfcbVerified TyfcbStatus = "verified"
	TyfcbRejected TyfcbStatus = "rejected"
	TyfcbVoid     TyfcbStatus = "void"
)

func ValidTyfcbStatus(v string) bool {
	switch TyfcbStatus(v) {
	case TyfcbPending, TyfcbVerified, TyfcbRejected:
		return true
	}
	return false
}

type VisitorStatus string

const (
	VisitorTerdaftar  VisitorStatus = "terdaftar"
	VisitorHadir      VisitorStatus = "hadir"
	VisitorHadirPenuh VisitorStatus = "hadir_penuh"
)

func ValidVisitorStatus(v string) bool {
	switch VisitorStatus(v) {
	case VisitorTerdaftar, VisitorHadir, VisitorHadirPenuh:
		return true
	}
	return false
}

type LedgerCategory string

const (
	CategoryTyfcb   LedgerCategory = "tyfcb"
	CategoryVisitor LedgerCategory = "visitor"
	CategoryBonus   LedgerCategory = "bonus"
)

// User is an authentication principal.
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     Role   `json:"role"`
}

type Season struct {
	ID       string     `json:"id"`
	Nama     string     `json:"nama"`
	StartsOn *time.Time `json:"starts_on"`
	EndsOn   *time.Time `json:"ends_on"`
	Status   string     `json:"status"`
}

type Team struct {
	ID          string `json:"id"`
	SeasonID    string `json:"season_id"`
	NamaTim     string `json:"nama_tim"`
	MemberCount int    `json:"member_count,omitempty"`
}

type Classification struct {
	ID   string `json:"id"`
	Nama string `json:"nama"`
}

// Member is a user's competition profile within one season.
type Member struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	SeasonID        string      `json:"season_id"`
	TeamID          *string     `json:"team_id"`
	KlasifikasiID   *string     `json:"klasifikasi_id"`
	ColorStatus     ColorStatus `json:"color_status"`
	IsActive        bool        `json:"is_active"`
	FullName        string      `json:"full_name"`
	Email           string      `json:"email"`
	Role            Role        `json:"role"`
	NamaTim         *string     `json:"nama_tim"`
	KlasifikasiNama *string     `json:"klasifikasi_nama"`
}

type TyfcbEntry struct {
	ID                     string      `json:"id"`
	SeasonID               string      `json:"season_id"`
	GiverID                string      `json:"giver_id"`
	ReceiverID             string      `json:"receiver_id"`
	Nilai                  float64     `json:"nilai"`
	Tanggal                time.Time   `json:"tanggal"`
	Status                 TyfcbStatus `json:"status"`
	ComputedScore          *int        `json:"computed_score"`
	PairOrdinal            *int        `json:"pair_ordinal"`
	EventMultiplierApplied *float64    `json:"event_multiplier_applied"`
	RejectionReason        *string     `json:"rejection_reason"`
	GiverName              string      `json:"giver_name,omitempty"`
	ReceiverName           string      `json:"receiver_name,omitempty"`
	CreatedAt              time.Time   `json:"created_at"`
}

type Visitor struct {
	ID              string        `json:"id"`
	SeasonID        string        `json:"season_id"`
	Nama            string        `json:"nama"`
	Kontak          string        `json:"kontak"`
	InviterID       string        `json:"inviter_id"`
	TanggalUndang   time.Time     `json:"tanggal_undang"`
	StatusHadir     VisitorStatus `json:"status_hadir"`
	IsConverted     bool          `json:"is_converted"`
	IsVoid          bool          `json:"is_void"`
	TanggalKonversi *time.Time    `json:"tanggal_konversi"`
	InviterName     string        `json:"inviter_name,omitempty"`
	NamaTim         *string       `json:"nama_tim,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

// LedgerEntry is append-only: corrections are written as new rows with the
// opposite sign, never as updates or deletes.
type LedgerEntry struct {
	ID         string         `json:"id"`
	SeasonID   string         `json:"season_id"`
	MemberID   *string        `json:"member_id"`
	TeamID     *string        `json:"team_id"`
	Kategori   LedgerCategory `json:"kategori"`
	Points     int            `json:"points"`
	SumberRef  *string        `json:"sumber_ref"`
	Keterangan *string        `json:"keterangan"`
	CreatedAt  time.Time      `json:"created_at"`
}

type BoosterEvent struct {
	ID              string    `json:"id"`
	SeasonID        string    `json:"season_id"`
	Judul           string    `json:"judul"`
	Deskripsi       *string   `json:"deskripsi"`
	TanggalMulai    time.Time `json:"tanggal_mulai"`
	TanggalBerakhir time.Time `json:"tanggal_berakhir"`
	Poin            int       `json:"poin"`
	Status          string    `json:"status"`
}

type Badge struct {
	BadgeCode string  `json:"badge_code"`
	Nama      string  `json:"nama"`
	Deskripsi string  `json:"deskripsi"`
	Ikon      *string `json:"ikon"`
	EarnedAt  *string `json:"earned_at,omitempty"`
}

// TeamScore is a read model: the leaderboard aggregate for one team.
type TeamScore struct {
	TeamID       string  `json:"team_id"`
	NamaTim      string  `json:"nama_tim"`
	ScoreOverall int     `json:"score_overall"`
	ScoreTyfcb   int     `json:"score_tyfcb"`
	ScoreVisitor int     `json:"score_visitor"`
	ScoreBonus   int     `json:"score_bonus"`
	NilaiTyfcb   float64 `json:"nilai_tyfcb"`
	CountTyfcb   int     `json:"count_tyfcb"`
	CountVisitor int     `json:"count_visitor"`
}

// MemberScore is the same aggregate for an individual.
type MemberScore struct {
	MemberID     string  `json:"member_id"`
	FullName     string  `json:"full_name"`
	NamaTim      *string `json:"nama_tim"`
	ScoreOverall int     `json:"score_overall"`
	ScoreTyfcb   int     `json:"score_tyfcb"`
	ScoreVisitor int     `json:"score_visitor"`
	ScoreBonus   int     `json:"score_bonus"`
}

type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}
