package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

func tyfcbFixture(t *testing.T, status domain.TyfcbStatus, score int) (*Tyfcb, *fakeLedger, *fakeTyfcbRepo, *fakeTx) {
	t.Helper()

	entry := &domain.TyfcbEntry{
		ID:            "entry-1",
		SeasonID:      testSeasonID,
		GiverID:       "member-buyer",
		ReceiverID:    "member-seller",
		Nilai:         18_000_000,
		Status:        status,
		ComputedScore: &score,
	}

	repo := &fakeTyfcbRepo{entries: map[string]*domain.TyfcbEntry{"entry-1": entry}}
	ledger := &fakeLedger{}
	tx := &fakeTx{}
	members := &fakeMembers{byID: map[string]*domain.Member{
		"member-buyer":  member("member-buyer", "team-1"),
		"member-seller": member("member-seller", "team-2"),
	}}

	uc := NewTyfcb(repo, members, ledger, &fakeSeasons{season: season()},
		&fakeEvents{}, &fakeSpheres{}, NewBadges(&fakeBadges{}), tx)

	return uc, ledger, repo, tx
}

func TestVerifyCreditsTheGiver(t *testing.T) {
	uc, ledger, repo, tx := tyfcbFixture(t, domain.TyfcbPending, 80)

	if err := uc.SetStatus(context.Background(), "entry-1", domain.TyfcbVerified, "admin"); err != nil {
		t.Fatalf("verify: %v", err)
	}

	if len(ledger.entries) != 1 {
		t.Fatalf("wrote %d ledger rows, want 1", len(ledger.entries))
	}

	row := ledger.entries[0]
	if row.Points != 80 {
		t.Errorf("credited %d, want 80", row.Points)
	}
	// Points belong to the buyer, who earns them — not the member who filed.
	if row.MemberID == nil || *row.MemberID != "member-buyer" {
		t.Errorf("credited the wrong member: %v", row.MemberID)
	}
	// The bonus lands on the earner's team, not the filer's.
	if row.TeamID == nil || *row.TeamID != "team-1" {
		t.Errorf("credited the wrong team: %v", row.TeamID)
	}
	if repo.entries["entry-1"].Status != domain.TyfcbVerified {
		t.Error("status was not updated")
	}
	// The ledger write and the status change must share one transaction.
	if tx.depth != 1 {
		t.Errorf("used %d transactions, want 1", tx.depth)
	}
}

func TestRejectingAVerifiedEntryReversesIt(t *testing.T) {
	uc, ledger, _, _ := tyfcbFixture(t, domain.TyfcbVerified, 80)

	if err := uc.SetStatus(context.Background(), "entry-1", domain.TyfcbRejected, "admin"); err != nil {
		t.Fatalf("reject: %v", err)
	}

	if len(ledger.entries) != 1 || ledger.entries[0].Points != -80 {
		t.Fatalf("expected a single -80 reversal, got %+v", ledger.entries)
	}
}

// Verify then reject must leave the member exactly where they started, and the
// ledger must keep both rows rather than deleting the first.
func TestVerifyThenRejectNetsToZero(t *testing.T) {
	uc, ledger, _, _ := tyfcbFixture(t, domain.TyfcbPending, 80)
	ctx := context.Background()

	if err := uc.SetStatus(ctx, "entry-1", domain.TyfcbVerified, "admin"); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := uc.SetStatus(ctx, "entry-1", domain.TyfcbRejected, "admin"); err != nil {
		t.Fatalf("reject: %v", err)
	}

	if ledger.total() != 0 {
		t.Errorf("net %d points, want 0", ledger.total())
	}
	if len(ledger.entries) != 2 {
		t.Errorf("kept %d rows, want 2 — the ledger is append-only", len(ledger.entries))
	}
}

// Moving between two non-verified states must not touch the ledger at all.
func TestPendingToRejectedWritesNothing(t *testing.T) {
	uc, ledger, _, _ := tyfcbFixture(t, domain.TyfcbPending, 80)

	if err := uc.SetStatus(context.Background(), "entry-1", domain.TyfcbRejected, "admin"); err != nil {
		t.Fatalf("reject: %v", err)
	}

	if len(ledger.entries) != 0 {
		t.Errorf("wrote %d rows, want none", len(ledger.entries))
	}
}

func TestSetStatusRejectsNoOpsAndUnknownStates(t *testing.T) {
	cases := []struct {
		name    string
		from    domain.TyfcbStatus
		to      domain.TyfcbStatus
		wantErr error
	}{
		{"same status", domain.TyfcbPending, domain.TyfcbPending, domain.ErrInvalidInput},
		{"unknown status", domain.TyfcbPending, "bogus", domain.ErrInvalidInput},
		{"already void", domain.TyfcbVoid, domain.TyfcbVerified, domain.ErrConflict},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			uc, ledger, _, _ := tyfcbFixture(t, c.from, 80)

			err := uc.SetStatus(context.Background(), "entry-1", c.to, "admin")
			if !errors.Is(err, c.wantErr) {
				t.Errorf("got %v, want %v", err, c.wantErr)
			}
			if len(ledger.entries) != 0 {
				t.Error("a rejected transition must not touch the ledger")
			}
		})
	}
}

func TestVoidingAVerifiedEntryReturnsThePoints(t *testing.T) {
	uc, ledger, repo, _ := tyfcbFixture(t, domain.TyfcbVerified, 120)

	if err := uc.Void(context.Background(), "entry-1", "captain"); err != nil {
		t.Fatalf("void: %v", err)
	}

	if ledger.total() != -120 {
		t.Errorf("returned %d, want -120", ledger.total())
	}
	if repo.entries["entry-1"].Status != domain.TyfcbVoid {
		t.Error("entry was not marked void")
	}
}

// A pending entry never earned anything, so voiding it must not hand back
// points that were never given.
func TestVoidingAPendingEntryWritesNothing(t *testing.T) {
	uc, ledger, _, _ := tyfcbFixture(t, domain.TyfcbPending, 120)

	if err := uc.Void(context.Background(), "entry-1", "captain"); err != nil {
		t.Fatalf("void: %v", err)
	}

	if len(ledger.entries) != 0 {
		t.Errorf("wrote %d rows, want none", len(ledger.entries))
	}
}

func TestSubmitAppliesBandPairPenaltyAndEvent(t *testing.T) {
	repo := &fakeTyfcbRepo{pairCount: 1} // one prior transaction, so this is the second
	members := &fakeMembers{byID: map[string]*domain.Member{
		"buyer":  member("buyer", "team-1"),
		"seller": member("seller", "team-1"),
	}}

	uc := NewTyfcb(repo, members, &fakeLedger{}, &fakeSeasons{season: season()},
		&fakeEvents{event: &domain.WeeklyEvent{EventCode: domain.EventFounder}},
		&fakeSpheres{}, NewBadges(&fakeBadges{}), &fakeTx{})

	entry, err := uc.Submit(context.Background(), SubmitTyfcbInput{
		BuyerID:  "buyer",
		SellerID: "seller",
		Nilai:    18_000_000,
		Tanggal:  time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
	}, nil)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	// Band 80 × pair penalty ½ × Founder 1.5 = 60.
	if entry.ComputedScore == nil || *entry.ComputedScore != 60 {
		t.Errorf("scored %v, want 60", entry.ComputedScore)
	}
	if entry.PairOrdinal == nil || *entry.PairOrdinal != 2 {
		t.Errorf("pair ordinal %v, want 2", entry.PairOrdinal)
	}
	// A submission is never scored into the ledger until an admin verifies it.
	if entry.Status != domain.TyfcbPending {
		t.Errorf("status %s, want pending", entry.Status)
	}
}

func TestSubmitRefusesSelfDealing(t *testing.T) {
	uc := NewTyfcb(&fakeTyfcbRepo{}, &fakeMembers{byID: map[string]*domain.Member{}},
		&fakeLedger{}, &fakeSeasons{season: season()}, &fakeEvents{},
		&fakeSpheres{}, NewBadges(&fakeBadges{}), &fakeTx{})

	_, err := uc.Submit(context.Background(), SubmitTyfcbInput{
		BuyerID:  "same",
		SellerID: "same",
		Nilai:    1000,
		Tanggal:  time.Now(),
	}, nil)

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("got %v, want an invalid-input error", err)
	}
}
