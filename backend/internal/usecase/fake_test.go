package usecase

import (
	"context"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// In-memory doubles for the repository contracts. They exist so the use cases
// can be tested for what they orchestrate — which ledger rows get written, in
// what order, under which transaction — without a database in the loop.

const testSeasonID = "season-1"

type fakeTx struct{ depth int }

// WithinTx records nesting so a test can assert that work happened inside one
// transaction rather than several.
func (f *fakeTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	f.depth++
	return fn(ctx)
}

type fakeSeasons struct{ season *domain.Season }

func (f *fakeSeasons) FindActive(context.Context) (*domain.Season, error) { return f.season, nil }
func (f *fakeSeasons) FindByName(context.Context, string) (*domain.Season, error) {
	return f.season, nil
}

type fakeMembers struct{ byID map[string]*domain.Member }

func (f *fakeMembers) FindByID(_ context.Context, id string) (*domain.Member, error) {
	return f.byID[id], nil
}
func (f *fakeMembers) FindByUserAndSeason(context.Context, string, string) (*domain.Member, error) {
	return nil, nil
}
func (f *fakeMembers) ListBySeason(context.Context, string) ([]domain.Member, error) {
	out := []domain.Member{}
	for _, m := range f.byID {
		out = append(out, *m)
	}
	return out, nil
}
func (f *fakeMembers) ListByTeam(context.Context, string) ([]domain.Member, error) { return nil, nil }
func (f *fakeMembers) ListFiltered(ctx context.Context, _ domain.MemberFilter) ([]domain.Member, int, error) {
	members, err := f.ListBySeason(ctx, "")
	return members, len(members), err
}
func (f *fakeMembers) Search(context.Context, string, string, int) ([]domain.Member, error) {
	return nil, nil
}
func (f *fakeMembers) Create(context.Context, *domain.Member) (string, error) { return "", nil }
func (f *fakeMembers) Update(context.Context, *domain.Member) error           { return nil }
func (f *fakeMembers) CountBySeason(context.Context, string) (int, error)     { return 0, nil }

// fakeLedger records every append so a test can assert on the exact rows a use
// case produced, including their sign and source key.
type fakeLedger struct {
	entries []domain.LedgerEntry
}

func (f *fakeLedger) Append(_ context.Context, e *domain.LedgerEntry) error {
	f.entries = append(f.entries, *e)
	return nil
}
func (f *fakeLedger) TeamScores(context.Context, string) ([]domain.TeamScore, error) {
	return nil, nil
}
func (f *fakeLedger) MemberScores(context.Context, string, int) ([]domain.MemberScore, error) {
	return nil, nil
}
func (f *fakeLedger) MemberScore(context.Context, string, string) (*domain.MemberScore, error) {
	return &domain.MemberScore{}, nil
}
func (f *fakeLedger) TeamHistory(context.Context, string, string, string) ([]domain.LedgerEntry, error) {
	return nil, nil
}

// SumBySource mirrors what the real ledger would report for a source key.
func (f *fakeLedger) SumBySource(_ context.Context, ref string) (int, error) {
	total := 0
	for _, e := range f.entries {
		if e.SumberRef != nil && *e.SumberRef == ref {
			total += e.Points
		}
	}
	return total, nil
}

// total sums every recorded row, which is the leaderboard a member would see.
func (f *fakeLedger) total() int {
	sum := 0
	for _, e := range f.entries {
		sum += e.Points
	}
	return sum
}

type fakeTyfcbRepo struct {
	entries   map[string]*domain.TyfcbEntry
	pairCount int
	created   []domain.TyfcbEntry
}

func (f *fakeTyfcbRepo) FindByID(_ context.Context, id string) (*domain.TyfcbEntry, error) {
	return f.entries[id], nil
}
func (f *fakeTyfcbRepo) List(context.Context, domain.TyfcbFilter) ([]domain.TyfcbEntry, error) {
	return nil, nil
}
func (f *fakeTyfcbRepo) ListPaged(context.Context, domain.TyfcbFilter) ([]domain.TyfcbEntry, int, error) {
	return nil, 0, nil
}
func (f *fakeTyfcbRepo) CountPair(context.Context, string, string, string) (int, error) {
	return f.pairCount, nil
}
func (f *fakeTyfcbRepo) Create(_ context.Context, e *domain.TyfcbEntry, _ *string) (string, error) {
	f.created = append(f.created, *e)
	return "entry-new", nil
}
func (f *fakeTyfcbRepo) UpdateStatus(_ context.Context, id string, status domain.TyfcbStatus, _ *string, _ *time.Time) error {
	if e, ok := f.entries[id]; ok {
		e.Status = status
	}
	return nil
}
func (f *fakeTyfcbRepo) Void(_ context.Context, id, _ string) error {
	if e, ok := f.entries[id]; ok {
		e.Status = domain.TyfcbVoid
	}
	return nil
}
func (f *fakeTyfcbRepo) CountByStatus(context.Context, string) (map[string]int, error) {
	return nil, nil
}

type fakeVisitorRepo struct {
	visitors map[string]*domain.Visitor
	// guardFails makes the next guarded update lose, simulating a concurrent
	// writer that already moved the row.
	guardFails bool
}

func (f *fakeVisitorRepo) FindByID(_ context.Context, id string) (*domain.Visitor, error) {
	return f.visitors[id], nil
}
func (f *fakeVisitorRepo) List(context.Context, domain.VisitorFilter) ([]domain.Visitor, error) {
	return nil, nil
}
func (f *fakeVisitorRepo) ListPaged(context.Context, domain.VisitorFilter) ([]domain.Visitor, int, error) {
	return nil, 0, nil
}
func (f *fakeVisitorRepo) Create(context.Context, *domain.Visitor, *string) (string, error) {
	return "visitor-new", nil
}
func (f *fakeVisitorRepo) UpdateStatusGuarded(_ context.Context, id string, from, to domain.VisitorStatus) (bool, error) {
	if f.guardFails {
		return false, nil
	}
	v, ok := f.visitors[id]
	if !ok || v.StatusHadir != from {
		return false, nil
	}
	v.StatusHadir = to
	return true, nil
}
func (f *fakeVisitorRepo) UpdateConversionGuarded(_ context.Context, id string, from, to bool) (bool, error) {
	if f.guardFails {
		return false, nil
	}
	v, ok := f.visitors[id]
	if !ok || v.IsConverted != from {
		return false, nil
	}
	v.IsConverted = to
	return true, nil
}
func (f *fakeVisitorRepo) Void(_ context.Context, id, _ string) error {
	if v, ok := f.visitors[id]; ok {
		v.IsVoid = true
	}
	return nil
}

// fakeEvents returns one scheduled event for every date, or none.
type fakeEvents struct{ event *domain.WeeklyEvent }

func (f *fakeEvents) ActiveOn(context.Context, string, time.Time) (*domain.WeeklyEvent, error) {
	return f.event, nil
}
func (f *fakeEvents) ListBySeason(context.Context, string) ([]domain.WeeklyEvent, error) {
	return nil, nil
}
func (f *fakeEvents) Upsert(context.Context, *domain.WeeklyEvent) (string, error) { return "", nil }
func (f *fakeEvents) Delete(context.Context, string) error                        { return nil }

type fakeSpheres struct{ shares bool }

func (f *fakeSpheres) ListBySeason(context.Context, string) ([]domain.ContactSphere, error) {
	return nil, nil
}
func (f *fakeSpheres) Create(context.Context, string, string, *string) (string, error) {
	return "", nil
}
func (f *fakeSpheres) Delete(context.Context, string) error               { return nil }
func (f *fakeSpheres) SetMembers(context.Context, string, []string) error { return nil }
func (f *fakeSpheres) SharesSphere(context.Context, string, *string, *string) (bool, error) {
	return f.shares, nil
}

// fakeBadges records which badges were offered, so a test can assert that
// evaluation ran without asserting on the rules themselves.
type fakeBadges struct {
	stats   domain.BadgeStats
	awarded []string
}

func (f *fakeBadges) List(context.Context) ([]domain.Badge, error) { return nil, nil }
func (f *fakeBadges) ListForMember(context.Context, string) ([]domain.Badge, error) {
	return nil, nil
}
func (f *fakeBadges) Award(_ context.Context, _, code string) error {
	f.awarded = append(f.awarded, code)
	return nil
}
func (f *fakeBadges) Stats(context.Context, string, string) (domain.BadgeStats, error) {
	return f.stats, nil
}

// ── fixtures ──────────────────────────────────────────────────────────────

func member(id, teamID string) *domain.Member {
	return &domain.Member{
		ID:          id,
		SeasonID:    testSeasonID,
		TeamID:      &teamID,
		ColorStatus: domain.ColorHijau,
		IsActive:    true,
	}
}

func season() *domain.Season {
	return &domain.Season{ID: testSeasonID, Nama: "BRAG 2026", Status: "active"}
}

func ptr[T any](v T) *T { return &v }
