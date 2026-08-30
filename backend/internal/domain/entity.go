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

type RaffleSource string

const (
	RaffleFromScore     RaffleSource = "score"
	RaffleFromVisitor   RaffleSource = "visitor"
	RaffleFromTyfcbPair RaffleSource = "tyfcb_pair"
)

// Prize is one item in the pool, either seeded by the committee or donated by
// a member. Donations start pending until an admin approves them.
type Prize struct {
	ID             string   `json:"id"`
	SeasonID       string   `json:"season_id"`
	NamaHadiah     string   `json:"nama_hadiah"`
	Deskripsi      *string  `json:"deskripsi"`
	NilaiEstimasi  *float64 `json:"nilai_estimasi"`
	DonaturID      *string  `json:"donatur_id"`
	DonaturNama    *string  `json:"donatur_nama,omitempty"`
	Alokasi        string   `json:"alokasi"`
	KategoriTarget *string  `json:"kategori_target"`
	Status         string   `json:"status"`
	PemenangID     *string  `json:"pemenang_id"`
	PemenangNama   *string  `json:"pemenang_nama,omitempty"`
}

func ValidPrizeAlokasi(v string) bool { return v == "kategori" || v == "undian" }

func ValidPrizeStatus(v string) bool {
	switch v {
	case "pending", "approved", "rejected", "awarded":
		return true
	}
	return false
}

// ActivityItem is one entry in the season-wide feed that backs both the
// activity page and the notification bell.
type ActivityItem struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"` // "tyfcb" | "visitor"
	ActorName  string    `json:"actor_name"`
	TargetName string    `json:"target_name"`
	Amount     *float64  `json:"amount,omitempty"`
	Status     string    `json:"status"`
	Points     *int      `json:"points,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// ContactSphere is a named set of classifications that naturally refer each
// other business — a wedding sphere might hold Photography, Catering, Venue.
type ContactSphere struct {
	ID          string           `json:"id"`
	SeasonID    string           `json:"season_id"`
	Nama        string           `json:"nama"`
	Deskripsi   *string          `json:"deskripsi"`
	Klasifikasi []Classification `json:"klasifikasi"`
}

// OneToOne is a recorded meeting between two members. The pair is stored in a
// canonical order so a duplicate is caught whichever side files it.
type OneToOne struct {
	ID          string    `json:"id"`
	SeasonID    string    `json:"season_id"`
	MemberA     string    `json:"member_a"`
	MemberB     string    `json:"member_b"`
	MemberAName string    `json:"member_a_name,omitempty"`
	MemberBName string    `json:"member_b_name,omitempty"`
	Tanggal     time.Time `json:"tanggal"`
	Catatan     *string   `json:"catatan"`
	CreatedAt   time.Time `json:"created_at"`
}

// TyfcbStatusChange is one moderation decision. It travels as a value because
// the fields go together: who decided, when, and — for a rejection — why.
type TyfcbStatusChange struct {
	From, To   TyfcbStatus
	VerifiedBy *string
	VerifiedAt *time.Time
	// Reason is set when rejecting and cleared otherwise, so an entry that is
	// rejected and later approved does not keep an explanation that no longer
	// describes it.
	Reason *string
}

// APIKey is machine access to the API. It authenticates as a user and
// inherits that user's role, so the route guards need no second permission
// model — a key can never reach anything its owner could not.
type APIKey struct {
	ID   string `json:"id"`
	Nama string `json:"nama"`
	// Prefix is the visible start of the key, kept so a person can tell their
	// keys apart. It identifies; it does not open anything.
	Prefix     string     `json:"prefix"`
	UserID     string     `json:"user_id"`
	UserName   string     `json:"user_name,omitempty"`
	UserEmail  string     `json:"user_email,omitempty"`
	ReadOnly   bool       `json:"read_only"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Active reports whether the key may still be used.
func (k *APIKey) Active(now time.Time) bool {
	if k.RevokedAt != nil {
		return false
	}
	return k.ExpiresAt == nil || k.ExpiresAt.After(now)
}

// Status describes a key for the management screen in one word.
func (k *APIKey) Status(now time.Time) string {
	switch {
	case k.RevokedAt != nil:
		return "revoked"
	case k.ExpiresAt != nil && !k.ExpiresAt.After(now):
		return "expired"
	default:
		return "active"
	}
}
