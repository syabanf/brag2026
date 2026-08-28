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
}

type SeasonRepository interface {
	FindActive(ctx context.Context) (*Season, error)
	FindByName(ctx context.Context, nama string) (*Season, error)
}

type MemberRepository interface {
	FindByUserAndSeason(ctx context.Context, userID, seasonID string) (*Member, error)
	FindByID(ctx context.Context, id string) (*Member, error)
	ListBySeason(ctx context.Context, seasonID string) ([]Member, error)
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
	Limit      int
}

type TyfcbRepository interface {
	FindByID(ctx context.Context, id string) (*TyfcbEntry, error)
	List(ctx context.Context, f TyfcbFilter) ([]TyfcbEntry, error)
	CountPair(ctx context.Context, giverID, receiverID, seasonID string) (int, error)
	Create(ctx context.Context, e *TyfcbEntry, submittedBy *string) (string, error)
	UpdateStatus(ctx context.Context, id string, status TyfcbStatus, verifiedBy *string, verifiedAt *time.Time) error
	Void(ctx context.Context, id, voidedBy string) error
	CountByStatus(ctx context.Context, seasonID string) (map[string]int, error)
}

type VisitorFilter struct {
	SeasonID  string
	Status    string
	InviterID string
	TeamID    string
	Limit     int
}

type VisitorRepository interface {
	FindByID(ctx context.Context, id string) (*Visitor, error)
	List(ctx context.Context, f VisitorFilter) ([]Visitor, error)
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
	MemberScores(ctx context.Context, seasonID string, limit int) ([]MemberScore, error)
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

type PrizeRepository interface {
	List(ctx context.Context, seasonID string, status string) ([]Prize, error)
	FindByID(ctx context.Context, id string) (*Prize, error)
	Create(ctx context.Context, p *Prize) (string, error)
	SetStatus(ctx context.Context, id, status string, pemenangID *string) error
	Delete(ctx context.Context, id string) error
	CountApprovedDonations(ctx context.Context, memberID string) (int, error)

	// Raffle tickets are rebuilt from scratch per member, so re-running the
	// issue pass is idempotent.
	ReplaceTickets(ctx context.Context, seasonID, memberID string, bySource map[RaffleSource]int) error
	TicketCounts(ctx context.Context, seasonID string) (map[string]int, error)
	RaffleInputs(ctx context.Context, seasonID, memberID string) (score, visitors, newPairs int, err error)
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
