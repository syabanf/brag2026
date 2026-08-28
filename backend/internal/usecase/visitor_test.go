package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

func visitorFixture(t *testing.T, status domain.VisitorStatus, converted bool, event *domain.WeeklyEvent) (*Visitor, *fakeLedger, *fakeVisitorRepo) {
	t.Helper()

	visitor := &domain.Visitor{
		ID:          "visitor-1",
		SeasonID:    testSeasonID,
		InviterID:   "member-1",
		StatusHadir: status,
		IsConverted: converted,
	}

	repo := &fakeVisitorRepo{visitors: map[string]*domain.Visitor{"visitor-1": visitor}}
	ledger := &fakeLedger{}
	members := &fakeMembers{byID: map[string]*domain.Member{"member-1": member("member-1", "team-1")}}

	uc := NewVisitor(repo, members, ledger, &fakeEvents{event: event}, NewBadges(&fakeBadges{}), &fakeTx{})
	return uc, ledger, repo
}

func TestAttendanceMilestonesAwardTheDifference(t *testing.T) {
	cases := []struct {
		name string
		from domain.VisitorStatus
		to   domain.VisitorStatus
		want int
	}{
		{"registered to attended", domain.VisitorTerdaftar, domain.VisitorHadir, 20},
		{"attended to fully attended", domain.VisitorHadir, domain.VisitorHadirPenuh, 30},
		{"straight to fully attended", domain.VisitorTerdaftar, domain.VisitorHadirPenuh, 50},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			uc, ledger, _ := visitorFixture(t, c.from, false, nil)

			to := string(c.to)
			if err := uc.Update(context.Background(), "visitor-1", UpdateVisitorInput{StatusHadir: &to}); err != nil {
				t.Fatalf("update: %v", err)
			}

			if ledger.total() != c.want {
				t.Errorf("awarded %d, want %d", ledger.total(), c.want)
			}
		})
	}
}

// Walking the status up and back down must net to zero, or every admin
// correction would quietly inflate a leaderboard.
func TestAttendanceRoundTripNetsToZero(t *testing.T) {
	uc, ledger, _ := visitorFixture(t, domain.VisitorTerdaftar, false, nil)
	ctx := context.Background()

	for _, status := range []domain.VisitorStatus{
		domain.VisitorHadir, domain.VisitorHadirPenuh, domain.VisitorHadir, domain.VisitorTerdaftar,
	} {
		s := string(status)
		if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &s}); err != nil {
			t.Fatalf("to %s: %v", status, err)
		}
	}

	if ledger.total() != 0 {
		t.Errorf("round trip netted %d, want 0", ledger.total())
	}
}

// The regression this guards: a milestone awarded during a boosted week used
// to be reversed at the base rate, stranding the difference in the ledger.
func TestBoostedMilestoneIsReversedInFull(t *testing.T) {
	blitz := &domain.WeeklyEvent{EventCode: domain.EventVisitorBlitz}
	uc, ledger, _ := visitorFixture(t, domain.VisitorTerdaftar, false, blitz)
	ctx := context.Background()

	hadir := string(domain.VisitorHadir)
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &hadir}); err != nil {
		t.Fatalf("promote: %v", err)
	}

	// 20 × 1.5 = 30, not the base 20.
	if ledger.total() != 30 {
		t.Fatalf("boosted award was %d, want 30", ledger.total())
	}

	terdaftar := string(domain.VisitorTerdaftar)
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &terdaftar}); err != nil {
		t.Fatalf("demote: %v", err)
	}

	if ledger.total() != 0 {
		t.Errorf("after reversal the member kept %d points, want 0", ledger.total())
	}
}

func TestConversionBonusIsSeparateFromAttendance(t *testing.T) {
	// Built through the use case so the ledger matches the status, which is the
	// only way the pair can exist in production.
	uc, ledger, _ := visitorFixture(t, domain.VisitorTerdaftar, false, nil)
	ctx := context.Background()

	penuh := string(domain.VisitorHadirPenuh)
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &penuh}); err != nil {
		t.Fatalf("promote: %v", err)
	}

	converted := true
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{IsConverted: &converted}); err != nil {
		t.Fatalf("convert: %v", err)
	}
	if ledger.total() != 50+domain.ConversionPoints {
		t.Fatalf("earned %d, want %d", ledger.total(), 50+domain.ConversionPoints)
	}

	// Correcting attendance downwards must not claw back the conversion bonus,
	// which is why the two use separate ledger source keys.
	hadir := string(domain.VisitorHadir)
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &hadir}); err != nil {
		t.Fatalf("demote: %v", err)
	}

	want := 20 + domain.ConversionPoints
	if ledger.total() != want {
		t.Errorf("net %d, want %d — only the attendance step should move", ledger.total(), want)
	}
}

// A downgrade settles attendance back to the base value of the new status.
// Any boost earned on the steps being undone is given up with them, which is
// what keeps the ledger from drifting when an event is no longer running.
func TestDowngradeSettlesAttendanceToBaseRate(t *testing.T) {
	blitz := &domain.WeeklyEvent{EventCode: domain.EventVisitorBlitz}
	uc, ledger, _ := visitorFixture(t, domain.VisitorTerdaftar, false, blitz)
	ctx := context.Background()

	penuh := string(domain.VisitorHadirPenuh)
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &penuh}); err != nil {
		t.Fatalf("promote: %v", err)
	}
	// 50 × 1.5 = 75 while the blitz is running.
	if ledger.total() != 75 {
		t.Fatalf("boosted award %d, want 75", ledger.total())
	}

	hadir := string(domain.VisitorHadir)
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &hadir}); err != nil {
		t.Fatalf("demote: %v", err)
	}

	if ledger.total() != 20 {
		t.Errorf("after the correction the member holds %d, want the base 20", ledger.total())
	}
}

func TestVoidReturnsEverythingEarned(t *testing.T) {
	uc, ledger, repo := visitorFixture(t, domain.VisitorTerdaftar, false, nil)
	ctx := context.Background()

	penuh := string(domain.VisitorHadirPenuh)
	converted := true
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{
		StatusHadir: &penuh, IsConverted: &converted,
	}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if ledger.total() != 150 {
		t.Fatalf("earned %d, want 150", ledger.total())
	}

	if err := uc.Void(ctx, "visitor-1", "captain"); err != nil {
		t.Fatalf("void: %v", err)
	}

	if ledger.total() != 0 {
		t.Errorf("void left %d points behind, want 0", ledger.total())
	}
	if !repo.visitors["visitor-1"].IsVoid {
		t.Error("visitor was not marked void")
	}
}

// The guarded update is optimistic locking: when a concurrent writer has
// already moved the row, this request must lose rather than award again.
func TestConcurrentUpdateLosesWithoutAwarding(t *testing.T) {
	uc, ledger, repo := visitorFixture(t, domain.VisitorTerdaftar, false, nil)
	repo.guardFails = true

	hadir := string(domain.VisitorHadir)
	err := uc.Update(context.Background(), "visitor-1", UpdateVisitorInput{StatusHadir: &hadir})

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("got %v, want a conflict", err)
	}
	if len(ledger.entries) != 0 {
		t.Error("a losing update must not write points")
	}
}

func TestUpdateRejectsEmptyAndUnknownInput(t *testing.T) {
	uc, ledger, _ := visitorFixture(t, domain.VisitorTerdaftar, false, nil)
	ctx := context.Background()

	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("empty input: got %v, want invalid", err)
	}

	bogus := "not-a-status"
	if err := uc.Update(ctx, "visitor-1", UpdateVisitorInput{StatusHadir: &bogus}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("bad status: got %v, want invalid", err)
	}

	if len(ledger.entries) != 0 {
		t.Error("rejected input must not touch the ledger")
	}
}

func TestVoidedVisitorCannotBeUpdated(t *testing.T) {
	uc, _, repo := visitorFixture(t, domain.VisitorHadir, false, nil)
	repo.visitors["visitor-1"].IsVoid = true

	penuh := string(domain.VisitorHadirPenuh)
	err := uc.Update(context.Background(), "visitor-1", UpdateVisitorInput{StatusHadir: &penuh})

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("got %v, want a conflict", err)
	}
}
