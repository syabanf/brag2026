package domain

import (
	"context"
	"time"
)

// The interfaces below are owned by the domain and implemented by the outer
// persistence layer, so use cases depend on abstractions rather than pgx.

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, string, error) // user, password hash
	FindByID(ctx context.Context, id string) (*User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateRole(ctx context.Context, userID string, role Role) error
	Create(ctx context.Context, email, passwordHash, fullName string, role Role) (string, error)
	UpdateProfile(ctx context.Context, userID, fullName, email string) error
}

type SessionRepository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	FindUserByTokenHash(ctx context.Context, tokenHash string) (*User, error)
	Delete(ctx context.Context, tokenHash string) error
	DeleteAllForUser(ctx context.Context, userID string) error
	// DeleteExpired reclaims rows nobody can use any more; without it the
	// table grows for the life of the season.
	DeleteExpired(ctx context.Context) (int64, error)
}

type SeasonRepository interface {
	FindActive(ctx context.Context) (*Season, error)
	FindByName(ctx context.Context, nama string) (*Season, error)
}

type MemberRepository interface {
	FindByUserAndSeason(ctx context.Context, userID, seasonID string) (*Member, error)
	FindByID(ctx context.Context, id string) (*Member, error)
	ListBySeason(ctx context.Context, seasonID string) ([]Member, error)
	// ListFiltered is the paged, filtered roster the admin screen uses.
	ListFiltered(ctx context.Context, f MemberFilter) ([]Member, int, error)
	ListByTeam(ctx context.Context, teamID string) ([]Member, error)
	Search(ctx context.Context, seasonID, term string, limit int) ([]Member, error)
	Create(ctx context.Context, m *Member) (string, error)
	Update(ctx context.Context, m *Member) error
	CountBySeason(ctx context.Context, seasonID string) (int, error)
}

type TeamRepository interface {
	ListBySeason(ctx context.Context, seasonID string) ([]Team, error)
	FindByID(ctx context.Context, id string) (*Team, error)
	Create(ctx context.Context, seasonID, namaTim string) (string, error)
	Rename(ctx context.Context, id, namaTim string) error
	Delete(ctx context.Context, id string) error
}

type ClassificationRepository interface {
	List(ctx context.Context) ([]Classification, error)
	Create(ctx context.Context, nama string) (string, error)
	Rename(ctx context.Context, id, nama string) error
	Delete(ctx context.Context, id string) error
	CountMembers(ctx context.Context, id string) (int, error)
}

type TyfcbFilter struct {
	SeasonID   string
	Status     string
	GiverID    string
	ReceiverID string
	TeamID     string
	// Search matches either party's name, so an admin can find an entry
	// without knowing which side filed it.
	Search   string
	DateFrom *time.Time
	DateTo   *time.Time
	Limit    int
	Page     Page
}

type TyfcbRepository interface {
	FindByID(ctx context.Context, id string) (*TyfcbEntry, error)
	List(ctx context.Context, f TyfcbFilter) ([]TyfcbEntry, error)
	ListPaged(ctx context.Context, f TyfcbFilter) ([]TyfcbEntry, int, error)
	CountPair(ctx context.Context, giverID, receiverID, seasonID string) (int, error)
	Create(ctx context.Context, e *TyfcbEntry, submittedBy *string) (string, error)
	// UpdateStatusGuarded only succeeds when the row still holds `From`. Two
	// admins verifying the same entry would otherwise both credit it, and the
	// ledger has no way to take points back except another entry.
	UpdateStatusGuarded(ctx context.Context, id string, change TyfcbStatusChange) (bool, error)
	Void(ctx context.Context, id, voidedBy string) error
	CountByStatus(ctx context.Context, seasonID string) (map[string]int, error)
}

type VisitorFilter struct {
	SeasonID  string
	Status    string
	InviterID string
	TeamID    string
	// Search matches the guest name, their contact, or the inviter.
	Search    string
	Converted *bool
	Limit     int
	Page      Page
}

// MemberFilter narrows the admin roster, which is the longest list in the app.
type MemberFilter struct {
	SeasonID    string
	TeamID      string
	Role        string
	ColorStatus string
	IsActive    *bool
	// Search matches name or email.
	Search string
	Page   Page
}

type VisitorRepository interface {
	FindByID(ctx context.Context, id string) (*Visitor, error)
	List(ctx context.Context, f VisitorFilter) ([]Visitor, error)
	ListPaged(ctx context.Context, f VisitorFilter) ([]Visitor, int, error)
	Create(ctx context.Context, v *Visitor, submittedBy *string) (string, error)
	// UpdateStatusGuarded only succeeds when the row still holds `from`, which
	// keeps concurrent updates from awarding points twice.
	UpdateStatusGuarded(ctx context.Context, id string, from, to VisitorStatus) (bool, error)
	UpdateConversionGuarded(ctx context.Context, id string, from, to bool) (bool, error)
	Void(ctx context.Context, id, voidedBy string) error
}

type LedgerRepository interface {
	// Append is the only write: the ledger is never updated or deleted.
	Append(ctx context.Context, e *LedgerEntry) error
	TeamScores(ctx context.Context, seasonID string) ([]TeamScore, error)
	MemberScores(ctx context.Context, seasonID string, kategori ScoreCategory, limit int) ([]MemberScore, error)
	MemberScore(ctx context.Context, memberID, seasonID string) (*MemberScore, error)
	TeamHistory(ctx context.Context, teamID, seasonID, kategori string) ([]LedgerEntry, error)
	// SumBySource totals what a single source row has already been credited.
	// Reversals read this instead of recomputing, so a boosted award is always
	// given back in full even if the boost is no longer running.
	SumBySource(ctx context.Context, sumberRef string) (int, error)
}

type BoosterRepository interface {
	ListBySeason(ctx context.Context, seasonID string, activeOnly bool) ([]BoosterEvent, error)
	FindByID(ctx context.Context, id string) (*BoosterEvent, error)
	Create(ctx context.Context, b *BoosterEvent) (string, error)
	Update(ctx context.Context, b *BoosterEvent) error
	Delete(ctx context.Context, id string) error
}

type BadgeRepository interface {
	List(ctx context.Context) ([]Badge, error)
	ListForMember(ctx context.Context, memberID string) ([]Badge, error)
	// Award is idempotent, so the evaluator can re-offer a badge the member
	// already holds without special-casing it.
	Award(ctx context.Context, memberID, badgeCode string) error
	// Stats gathers everything the badge rules need in one round trip.
	Stats(ctx context.Context, memberID, seasonID string) (BadgeStats, error)
}

// ScoringPassRepository backs the periodic settlements: bonuses that depend on
// a whole day or week of activity rather than a single submission.
type ScoringPassRepository interface {
	// RosterForWeek reports, per team, how many active members there are and
	// how many of them scored in the window.
	RosterForWeek(ctx context.Context, seasonID string, from, to time.Time) ([]RosterStatus, error)
	// TopTyfcbOfDay returns the single largest verified TYFCB of a day, or nil.
	TopTyfcbOfDay(ctx context.Context, seasonID string, day time.Time) (*TyfcbEntry, error)
	// MembersWithScoringDays lists members who scored on at least `days`
	// distinct dates inside the window.
	MembersWithScoringDays(ctx context.Context, seasonID string, from, to time.Time, days int) ([]string, error)
	// AlreadySettled guards against paying the same pass twice: it looks for a
	// ledger row already written with this source key.
	AlreadySettled(ctx context.Context, sumberRef string) (bool, error)
	// TeamOf resolves a member's team so a flat bonus lands on the right board.
	TeamOf(ctx context.Context, memberID string) (*string, error)
}

// ActivityRepository serves the season-wide feed: TYFCB and visitor events
// merged into one reverse-chronological stream.
type ActivityRepository interface {
	Recent(ctx context.Context, seasonID string, limit int) ([]ActivityItem, error)
}

type ContactSphereRepository interface {
	ListBySeason(ctx context.Context, seasonID string) ([]ContactSphere, error)
	Create(ctx context.Context, seasonID, nama string, deskripsi *string) (string, error)
	Delete(ctx context.Context, id string) error
	SetMembers(ctx context.Context, sphereID string, klasifikasiIDs []string) error
	// SharesSphere reports whether two classifications sit in a common sphere,
	// which is what POWER_TEAM rewards.
	SharesSphere(ctx context.Context, seasonID string, a, b *string) (bool, error)
}

type OneToOneRepository interface {
	List(ctx context.Context, seasonID, memberID string, limit int) ([]OneToOne, error)
	Create(ctx context.Context, o *OneToOne, submittedBy *string) (string, error)
	Delete(ctx context.Context, id string) error
	// PairsWithTyfcbInWindow returns the member pairs that both met one-to-one
	// and closed verified business inside the window — the ONE_TO_ONE payoff.
	PairsWithTyfcbInWindow(ctx context.Context, seasonID string, from, to time.Time) ([][2]string, error)
}

// TicketCount is one member's raffle entitlement after a rebuild.
type TicketCount struct {
	MemberID string
	FullName string
	NamaTim  *string
	Tickets  int
}

type PrizeRepository interface {
	List(ctx context.Context, seasonID string, status string) ([]Prize, error)
	FindByID(ctx context.Context, id string) (*Prize, error)
	Create(ctx context.Context, p *Prize) (string, error)
	SetStatus(ctx context.Context, id, status string, pemenangID *string) error
	Delete(ctx context.Context, id string) error
	CountApprovedDonations(ctx context.Context, memberID string) (int, error)

	// RebuildTickets recomputes the whole season's entitlement in one pass.
	// Rebuilding rather than appending keeps it idempotent, and doing it set-
	// at-a-time avoids a query per member and a row per ticket.
	RebuildTickets(ctx context.Context, seasonID string) ([]TicketCount, error)
	TicketCounts(ctx context.Context, seasonID string) (map[string]int, error)

	// DrawWinner picks one ticket at random and claims the prize for its
	// holder in a single statement, guarded on the prize having no winner
	// yet. Two admins pressing Draw at the same moment therefore produce one
	// winner, not two, and the second is told the draw already happened.
	//
	// Selection is one row per ticket, so a member with four tickets is four
	// times as likely to be picked — which is the whole point of issuing them.
	DrawWinner(ctx context.Context, seasonID, prizeID string) (*Prize, error)
}

type WeeklyEventRepository interface {
	// ActiveOn returns the event covering a date, or nil when none is
	// scheduled — the multiplier then stays at 1.
	ActiveOn(ctx context.Context, seasonID string, day time.Time) (*WeeklyEvent, error)
	ListBySeason(ctx context.Context, seasonID string) ([]WeeklyEvent, error)
	Upsert(ctx context.Context, e *WeeklyEvent) (string, error)
	Delete(ctx context.Context, id string) error
}

// TxManager lets a use case make several repository calls atomic without
// knowing what a transaction is. Implementations put the handle on the
// context; repositories pick it up transparently.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
