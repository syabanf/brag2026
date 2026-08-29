package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// fakePrizeRepo stands in for the prize table. DrawWinner mirrors the real
// statement's guard: it writes only while the prize has no winner.
type fakePrizeRepo struct {
	prizes  map[string]*domain.Prize
	tickets []string
	// drawnBy records who the guarded statement handed the prize to, so a test
	// can tell a refusal apart from a silent no-op.
	draws int
}

func (f *fakePrizeRepo) List(context.Context, string, string) ([]domain.Prize, error) {
	return nil, nil
}
func (f *fakePrizeRepo) FindByID(_ context.Context, id string) (*domain.Prize, error) {
	p, ok := f.prizes[id]
	if !ok {
		return nil, nil
	}
	copied := *p
	return &copied, nil
}
func (f *fakePrizeRepo) Create(context.Context, *domain.Prize) (string, error) { return "", nil }
func (f *fakePrizeRepo) SetStatus(_ context.Context, id, status string, pemenangID *string) error {
	p, ok := f.prizes[id]
	if !ok {
		return domain.NotFound("Hadiah tidak ditemukan.")
	}
	p.Status = status
	if pemenangID != nil {
		p.PemenangID = pemenangID
	}
	return nil
}
func (f *fakePrizeRepo) Delete(context.Context, string) error { return nil }
func (f *fakePrizeRepo) CountApprovedDonations(context.Context, string) (int, error) {
	return 0, nil
}
func (f *fakePrizeRepo) RebuildTickets(context.Context, string) ([]domain.TicketCount, error) {
	return nil, nil
}
func (f *fakePrizeRepo) TicketCounts(context.Context, string) (map[string]int, error) {
	return nil, nil
}
func (f *fakePrizeRepo) DrawWinner(_ context.Context, seasonID, prizeID string) (*domain.Prize, error) {
	p, ok := f.prizes[prizeID]
	if !ok || p.SeasonID != seasonID || p.PemenangID != nil || len(f.tickets) == 0 {
		return nil, nil
	}
	f.draws++
	winner := f.tickets[0]
	p.PemenangID = &winner
	p.Status = "awarded"
	copied := *p
	return &copied, nil
}

func prizeFixture(t *testing.T, p *domain.Prize, tickets []string) (*Prize, *fakePrizeRepo) {
	t.Helper()

	repo := &fakePrizeRepo{
		prizes:  map[string]*domain.Prize{p.ID: p},
		tickets: tickets,
	}
	uc := NewPrize(repo, stressMembers(), &fakeSeasons{season: season()},
		NewBadges(&fakeBadges{}), &fakeTx{})
	return uc, repo
}

func undianPrize() *domain.Prize {
	return &domain.Prize{
		ID: "prize-1", SeasonID: testSeasonID,
		NamaHadiah: "Voucher makan", Alokasi: "undian", Status: "approved",
	}
}

func TestDrawPicksAWinnerAndAwardsThePrize(t *testing.T) {
	uc, _ := prizeFixture(t, undianPrize(), []string{"member-buyer"})

	won, err := uc.Draw(context.Background(), "prize-1")
	if err != nil {
		t.Fatalf("draw: %v", err)
	}
	if won.PemenangID == nil {
		t.Fatal("the prize came back with no winner")
	}
	if *won.PemenangID != "member-buyer" {
		t.Errorf("winner is %q", *won.PemenangID)
	}
	if won.Status != "awarded" {
		t.Errorf("status is %q, want awarded", won.Status)
	}
}

// The result is announced in the room, so a second press must not re-roll it.
func TestDrawRefusesToRunTwice(t *testing.T) {
	uc, repo := prizeFixture(t, undianPrize(), []string{"member-buyer", "member-seller"})
	ctx := context.Background()

	if _, err := uc.Draw(ctx, "prize-1"); err != nil {
		t.Fatalf("first draw: %v", err)
	}
	first := *repo.prizes["prize-1"].PemenangID

	_, err := uc.Draw(ctx, "prize-1")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second draw gave %v, want a conflict", err)
	}
	if got := *repo.prizes["prize-1"].PemenangID; got != first {
		t.Errorf("the winner changed from %q to %q", first, got)
	}
	if repo.draws != 1 {
		t.Errorf("the guarded statement ran %d times, want 1", repo.draws)
	}
}

func TestDrawRejectsPrizesThatAreNotDrawable(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*domain.Prize)
		wantErr error
	}{
		{"allocated by category", func(p *domain.Prize) { p.Alokasi = "kategori" }, domain.ErrInvalidInput},
		{"still pending approval", func(p *domain.Prize) { p.Status = "pending" }, domain.ErrInvalidInput},
		{"rejected donation", func(p *domain.Prize) { p.Status = "rejected" }, domain.ErrInvalidInput},
		{"belongs to another season", func(p *domain.Prize) { p.SeasonID = "season-2" }, domain.ErrNotFound},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			prize := undianPrize()
			c.mutate(prize)
			uc, repo := prizeFixture(t, prize, []string{"member-buyer"})

			_, err := uc.Draw(context.Background(), "prize-1")
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("got %v, want %v", err, c.wantErr)
			}
			if repo.draws != 0 {
				t.Error("the draw ran despite the refusal")
			}
		})
	}
}

func TestDrawWithoutTicketsSaysSo(t *testing.T) {
	uc, _ := prizeFixture(t, undianPrize(), nil)

	_, err := uc.Draw(context.Background(), "prize-1")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("got %v, want an invalid-input error", err)
	}
	// The admin needs to know the fix is to issue tickets, not to retry.
	if err.Error() == "" || !contains(err.Error(), "tiket") {
		t.Errorf("the message does not mention tickets: %q", err)
	}
}

func TestDrawOnAMissingPrize(t *testing.T) {
	uc, _ := prizeFixture(t, undianPrize(), []string{"member-buyer"})

	if _, err := uc.Draw(context.Background(), "nope"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("got %v, want not-found", err)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
