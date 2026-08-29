package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// These tests put the scoring paths under concurrent load. The ledger is
// append-only, so a race that credits twice is not a transient glitch — the
// points stay on the board until someone notices and posts a correction by
// hand. Run with -race.

// ── thread-safe doubles ───────────────────────────────────────────────────

// syncLedger is the append-only ledger with a lock, since real appends are
// serialised by the database.
type syncLedger struct {
	mu      sync.Mutex
	entries []domain.LedgerEntry
}

func (l *syncLedger) Append(_ context.Context, e *domain.LedgerEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, *e)
	return nil
}
func (l *syncLedger) TeamScores(context.Context, string) ([]domain.TeamScore, error) {
	return nil, nil
}
func (l *syncLedger) MemberScores(context.Context, string, domain.ScoreCategory, int) ([]domain.MemberScore, error) {
	return nil, nil
}
func (l *syncLedger) MemberScore(context.Context, string, string) (*domain.MemberScore, error) {
	return &domain.MemberScore{}, nil
}
func (l *syncLedger) TeamHistory(context.Context, string, string, string) ([]domain.LedgerEntry, error) {
	return nil, nil
}
func (l *syncLedger) SumBySource(_ context.Context, ref string) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	total := 0
	for _, e := range l.entries {
		if e.SumberRef != nil && *e.SumberRef == ref {
			total += e.Points
		}
	}
	return total, nil
}
func (l *syncLedger) total() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	sum := 0
	for _, e := range l.entries {
		sum += e.Points
	}
	return sum
}
func (l *syncLedger) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// syncTyfcbRepo mirrors the real repository's locking: each statement is
// atomic on its own, but nothing holds a lock across the read and the write.
// That gap is exactly where a concurrent verification can slip through.
type syncTyfcbRepo struct {
	mu     sync.Mutex
	entry  domain.TyfcbEntry
	writes int
	// readBarrier, when set, holds every reader until the expected number have
	// arrived. Both callers then carry the same pre-transaction view, which is
	// the window a stale read actually lives in — reproducing it on purpose
	// beats hoping the scheduler produces it.
	readBarrier *barrier
	// staleRead, when set, is the status FindByID reports regardless of what
	// the row holds. It stands in for a caller whose read happened before
	// somebody else's write landed.
	staleRead *domain.TyfcbStatus
}

// barrier releases everyone once n participants have arrived.
type barrier struct {
	mu      sync.Mutex
	cond    *sync.Cond
	n       int
	arrived int
}

func newBarrier(n int) *barrier {
	b := &barrier{n: n}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *barrier) wait() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.arrived++
	if b.arrived >= b.n {
		b.cond.Broadcast()
		return
	}
	for b.arrived < b.n {
		b.cond.Wait()
	}
}

func (r *syncTyfcbRepo) FindByID(_ context.Context, id string) (*domain.TyfcbEntry, error) {
	r.mu.Lock()
	if id != r.entry.ID {
		r.mu.Unlock()
		return nil, nil
	}
	copied := r.entry
	if r.staleRead != nil {
		copied.Status = *r.staleRead
	}
	b := r.readBarrier
	r.mu.Unlock()

	// Held after the copy, so everyone leaves with the same view.
	if b != nil {
		b.wait()
	}
	return &copied, nil
}
func (r *syncTyfcbRepo) List(context.Context, domain.TyfcbFilter) ([]domain.TyfcbEntry, error) {
	return nil, nil
}
func (r *syncTyfcbRepo) ListPaged(context.Context, domain.TyfcbFilter) ([]domain.TyfcbEntry, int, error) {
	return nil, 0, nil
}
func (r *syncTyfcbRepo) CountPair(context.Context, string, string, string) (int, error) {
	return 0, nil
}
func (r *syncTyfcbRepo) Create(context.Context, *domain.TyfcbEntry, *string) (string, error) {
	return "entry-new", nil
}

// UpdateStatusGuarded refuses the write when the row has already moved on,
// which is what `where id = $n and status = $m` does in Postgres.
func (r *syncTyfcbRepo) UpdateStatusGuarded(_ context.Context, id string, c domain.TyfcbStatusChange) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id != r.entry.ID || r.entry.Status != c.From {
		return false, nil
	}
	r.entry.Status = c.To
	r.writes++
	return true, nil
}
func (r *syncTyfcbRepo) Void(_ context.Context, id string, from domain.TyfcbStatus, _ string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id != r.entry.ID || r.entry.Status != from {
		return false, nil
	}
	r.entry.Status = domain.TyfcbVoid
	r.writes++
	return true, nil
}
func (r *syncTyfcbRepo) CountByStatus(context.Context, string) (map[string]int, error) {
	return nil, nil
}
func (r *syncTyfcbRepo) status() domain.TyfcbStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.entry.Status
}

// syncVisitorRepo does the same for visitors, where the guarded updates are
// already in place.
type syncVisitorRepo struct {
	mu      sync.Mutex
	visitor domain.Visitor
}

func (r *syncVisitorRepo) FindByID(_ context.Context, id string) (*domain.Visitor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id != r.visitor.ID {
		return nil, nil
	}
	copied := r.visitor
	return &copied, nil
}
func (r *syncVisitorRepo) List(context.Context, domain.VisitorFilter) ([]domain.Visitor, error) {
	return nil, nil
}
func (r *syncVisitorRepo) ListPaged(context.Context, domain.VisitorFilter) ([]domain.Visitor, int, error) {
	return nil, 0, nil
}
func (r *syncVisitorRepo) Create(context.Context, *domain.Visitor, *string) (string, error) {
	return "visitor-new", nil
}
func (r *syncVisitorRepo) UpdateStatusGuarded(_ context.Context, id string, from, to domain.VisitorStatus) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id != r.visitor.ID || r.visitor.StatusHadir != from || r.visitor.IsVoid {
		return false, nil
	}
	r.visitor.StatusHadir = to
	return true, nil
}
func (r *syncVisitorRepo) UpdateConversionGuarded(_ context.Context, id string, from, to bool) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id != r.visitor.ID || r.visitor.IsConverted != from || r.visitor.IsVoid {
		return false, nil
	}
	r.visitor.IsConverted = to
	return true, nil
}
func (r *syncVisitorRepo) Void(_ context.Context, id, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.visitor.IsVoid = true
	return nil
}

// syncTx runs one transaction at a time, which is what Postgres does to these
// particular transactions: every one of them opens by updating the same row,
// so the second blocks on the first's row lock until it commits.
//
// The gap that matters is left wide open. Each request reads its entry with
// FindByID *before* WithinTx, so two callers still start from the same stale
// view of the row — which is exactly where a missing guard lets both of them
// award the same points.
type syncTx struct{ mu sync.Mutex }

func (t *syncTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return fn(ctx)
}

type syncBadgeRepo struct{ mu sync.Mutex }

func (b *syncBadgeRepo) List(context.Context) ([]domain.Badge, error) { return nil, nil }
func (b *syncBadgeRepo) ListForMember(context.Context, string) ([]domain.Badge, error) {
	return nil, nil
}
func (b *syncBadgeRepo) Award(context.Context, string, string) error { return nil }
func (b *syncBadgeRepo) Stats(context.Context, string, string) (domain.BadgeStats, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return domain.BadgeStats{}, nil
}

func stressMembers() *fakeMembers {
	return &fakeMembers{byID: map[string]*domain.Member{
		"member-buyer":  member("member-buyer", "team-1"),
		"member-seller": member("member-seller", "team-2"),
	}}
}

// hammer runs fn from `workers` goroutines released at the same moment, and
// returns how many succeeded. Starting them together is the point: staggered
// calls would never overlap in the window under test.
func hammer(workers int, fn func() error) (succeeded int, errs []error) {
	var start sync.WaitGroup
	var done sync.WaitGroup
	var mu sync.Mutex

	start.Add(1)
	done.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer done.Done()
			start.Wait()

			err := fn()

			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				succeeded++
			} else {
				errs = append(errs, err)
			}
		}()
	}

	start.Done()
	done.Wait()
	return succeeded, errs
}

// ── TYFCB ─────────────────────────────────────────────────────────────────

// Two admins clicking Verify on the same entry is an ordinary Monday, not an
// exotic race: the pending queue is shared and the entries are reviewed in
// bulk. Only one of them may move points.
func TestConcurrentVerifyCreditsOnce(t *testing.T) {
	const workers = 16
	const score = 80

	repo := &syncTyfcbRepo{entry: domain.TyfcbEntry{
		ID:            "entry-1",
		SeasonID:      testSeasonID,
		GiverID:       "member-buyer",
		ReceiverID:    "member-seller",
		Nilai:         18_000_000,
		Status:        domain.TyfcbPending,
		ComputedScore: ptr(score),
	}}
	ledger := &syncLedger{}

	uc := NewTyfcb(repo, stressMembers(), ledger, &fakeSeasons{season: season()},
		&fakeEvents{}, &fakeSpheres{}, NewBadges(&syncBadgeRepo{}), &syncTx{})

	ok, _ := hammer(workers, func() error {
		return uc.SetStatus(context.Background(), "entry-1", domain.TyfcbVerified, "admin", nil)
	})

	if ok != 1 {
		t.Errorf("%d of %d verifications were accepted, want 1", ok, workers)
	}
	if got := ledger.count(); got != 1 {
		t.Errorf("wrote %d ledger rows, want 1", got)
	}
	if got := ledger.total(); got != score {
		t.Errorf("credited %d points, want %d", got, score)
	}
	if got := repo.status(); got != domain.TyfcbVerified {
		t.Errorf("entry ended as %q, want verified", got)
	}
}

// Verify and void racing each other must not leave the entry both credited
// and voided.
//
// The barrier makes both requests read the entry as pending before either
// writes, which is the interleaving that matters: without a guard on the void,
// the verification credits the points and the void then closes the entry
// without reversing them, stranding the score on a written-off transaction.
func TestConcurrentVerifyAndVoidSettleToZeroOrCredit(t *testing.T) {
	const score = 120

	repo := &syncTyfcbRepo{
		entry: domain.TyfcbEntry{
			ID: "entry-1", SeasonID: testSeasonID,
			GiverID: "member-buyer", ReceiverID: "member-seller",
			Status: domain.TyfcbPending, ComputedScore: ptr(score),
		},
		readBarrier: newBarrier(2),
	}
	ledger := &syncLedger{}

	uc := NewTyfcb(repo, stressMembers(), ledger, &fakeSeasons{season: season()},
		&fakeEvents{}, &fakeSpheres{}, NewBadges(&syncBadgeRepo{}), &syncTx{})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = uc.SetStatus(context.Background(), "entry-1", domain.TyfcbVerified, "admin", nil)
	}()
	go func() { defer wg.Done(); _ = uc.Void(context.Background(), "entry-1", "admin") }()
	wg.Wait()

	// A voided entry is worth nothing; a verified one is worth its score.
	// Anything else means the two paths disagreed about what happened.
	total := ledger.total()
	if repo.status() == domain.TyfcbVoid {
		if total != 0 {
			t.Errorf("the entry is void but carries %d points", total)
		}
		return
	}
	if total != score {
		t.Errorf("the entry is %q but carries %d points, want %d", repo.status(), total, score)
	}
}

// Repeated verify/unverify must land exactly where it started. Any drift here
// compounds across a season.
func TestVerifyUnverifyCyclesNetToZero(t *testing.T) {
	const score = 65

	repo := &syncTyfcbRepo{entry: domain.TyfcbEntry{
		ID: "entry-1", SeasonID: testSeasonID,
		GiverID: "member-buyer", ReceiverID: "member-seller",
		Status: domain.TyfcbPending, ComputedScore: ptr(score),
	}}
	ledger := &syncLedger{}

	uc := NewTyfcb(repo, stressMembers(), ledger, &fakeSeasons{season: season()},
		&fakeEvents{}, &fakeSpheres{}, NewBadges(&syncBadgeRepo{}), &syncTx{})

	ctx := context.Background()
	for i := 0; i < 50; i++ {
		if err := uc.SetStatus(ctx, "entry-1", domain.TyfcbVerified, "admin", nil); err != nil {
			t.Fatalf("cycle %d verify: %v", i, err)
		}
		if err := uc.SetStatus(ctx, "entry-1", domain.TyfcbPending, "admin", nil); err != nil {
			t.Fatalf("cycle %d unverify: %v", i, err)
		}
	}

	if got := ledger.total(); got != 0 {
		t.Errorf("50 verify/unverify cycles drifted by %d points", got)
	}
	if got := ledger.count(); got != 100 {
		t.Errorf("wrote %d ledger rows, want 100 — corrections are appended, never erased", got)
	}
}

// ── Visitors ──────────────────────────────────────────────────────────────

// The visitor path already guards its updates. This pins that defence so it
// survives future edits.
func TestConcurrentVisitorPromotionAwardsOnce(t *testing.T) {
	const workers = 16

	repo := &syncVisitorRepo{visitor: domain.Visitor{
		ID: "visitor-1", SeasonID: testSeasonID,
		InviterID: "member-buyer", StatusHadir: domain.VisitorTerdaftar,
	}}
	ledger := &syncLedger{}

	uc := NewVisitor(repo, stressMembers(), ledger, &fakeEvents{},
		NewBadges(&syncBadgeRepo{}), &syncTx{})

	_, errs := hammer(workers, func() error {
		return uc.Update(context.Background(), "visitor-1", UpdateVisitorInput{
			StatusHadir: ptr(string(domain.VisitorHadirPenuh)),
		})
	})

	// Requests that arrive after the promotion has landed are no-ops and
	// report success — asking for a state the row already holds is not an
	// error. What must not happen is a second credit.
	if got := ledger.count(); got != 1 {
		t.Errorf("wrote %d ledger rows, want 1", got)
	}
	if got := ledger.total(); got != 50 {
		t.Errorf("awarded %d points, want 50", got)
	}
	// A request that did lose the race must be told why, not handed a
	// generic failure the UI cannot explain.
	for _, err := range errs {
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("a losing request got %v, want a conflict", err)
		}
	}
}

// Conversion is worth 100 on its own and races the same way.
func TestConcurrentConversionAwardsOnce(t *testing.T) {
	const workers = 12

	repo := &syncVisitorRepo{visitor: domain.Visitor{
		ID: "visitor-1", SeasonID: testSeasonID,
		InviterID: "member-buyer", StatusHadir: domain.VisitorHadirPenuh,
	}}
	ledger := &syncLedger{}

	uc := NewVisitor(repo, stressMembers(), ledger, &fakeEvents{},
		NewBadges(&syncBadgeRepo{}), &syncTx{})

	_, errs := hammer(workers, func() error {
		return uc.Update(context.Background(), "visitor-1", UpdateVisitorInput{
			IsConverted: ptr(true),
		})
	})

	for _, err := range errs {
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("a losing request got %v, want a conflict", err)
		}
	}
	if got := ledger.count(); got != 1 {
		t.Errorf("wrote %d ledger rows, want 1", got)
	}
	if got := ledger.total(); got != domain.ConversionPoints {
		t.Errorf("awarded %d points, want %d", got, domain.ConversionPoints)
	}
}

// Racing promotions to different statuses must settle on one of them, with
// points matching the status that actually stuck.
func TestConcurrentPromotionsToDifferentStatusesStayConsistent(t *testing.T) {
	repo := &syncVisitorRepo{visitor: domain.Visitor{
		ID: "visitor-1", SeasonID: testSeasonID,
		InviterID: "member-buyer", StatusHadir: domain.VisitorTerdaftar,
	}}
	ledger := &syncLedger{}

	uc := NewVisitor(repo, stressMembers(), ledger, &fakeEvents{},
		NewBadges(&syncBadgeRepo{}), &syncTx{})

	var wg sync.WaitGroup
	for _, status := range []domain.VisitorStatus{domain.VisitorHadir, domain.VisitorHadirPenuh} {
		wg.Add(1)
		go func(s domain.VisitorStatus) {
			defer wg.Done()
			_ = uc.Update(context.Background(), "visitor-1", UpdateVisitorInput{
				StatusHadir: ptr(string(s)),
			})
		}(status)
	}
	wg.Wait()

	final, _ := repo.FindByID(context.Background(), "visitor-1")
	want := domain.VisitorStatusDelta(domain.VisitorTerdaftar, final.StatusHadir)

	if got := ledger.total(); got != want {
		t.Errorf("the visitor is %q (worth %d) but carries %d points",
			final.StatusHadir, want, got)
	}
}

// A stale read reproduced exactly, with no scheduler involved: the caller
// decided to void an entry it believed was pending, and by the time its
// transaction runs the entry has been verified and credited. Acting on that
// stale view would close the entry without reversing the points — and the
// ledger is append-only, so nothing would ever take them back.
func TestVoidRefusesToActOnAStaleStatus(t *testing.T) {
	const score = 120

	repo := &syncTyfcbRepo{entry: domain.TyfcbEntry{
		ID: "entry-1", SeasonID: testSeasonID,
		GiverID: "member-buyer", ReceiverID: "member-seller",
		Status: domain.TyfcbPending, ComputedScore: ptr(score),
	}}
	ledger := &syncLedger{}

	uc := NewTyfcb(repo, stressMembers(), ledger, &fakeSeasons{season: season()},
		&fakeEvents{}, &fakeSpheres{}, NewBadges(&syncBadgeRepo{}), &syncTx{})

	ctx := context.Background()

	// Somebody else verifies it. This is the write the voider will not see.
	if err := uc.SetStatus(ctx, "entry-1", domain.TyfcbVerified, "admin", nil); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got := ledger.total(); got != score {
		t.Fatalf("setup credited %d, want %d", got, score)
	}

	// From here on every read reports the status the voider saw earlier,
	// while the row itself is verified.
	repo.staleRead = ptr(domain.TyfcbPending)

	err := uc.Void(ctx, "entry-1", "admin")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("void on a stale status returned %v, want a conflict", err)
	}

	repo.staleRead = nil
	if got := repo.status(); got != domain.TyfcbVerified {
		t.Errorf("the entry is %q, want it left verified", got)
	}
	if got := ledger.total(); got != score {
		t.Errorf("the ledger moved to %d; the refused void must write nothing", got)
	}
}
