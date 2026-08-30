package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// fakeAPIKeyRepo stores keys by digest, the way the real table does.
type fakeAPIKeyRepo struct {
	byID   map[string]*domain.APIKey
	byHash map[string]string // digest → id
	owners map[string]*domain.User
	touch  int
}

func newFakeAPIKeyRepo() *fakeAPIKeyRepo {
	return &fakeAPIKeyRepo{
		byID:   map[string]*domain.APIKey{},
		byHash: map[string]string{},
		owners: map[string]*domain.User{},
	}
}

func (f *fakeAPIKeyRepo) FindByHash(_ context.Context, hash string) (*domain.APIKey, *domain.User, error) {
	id, ok := f.byHash[hash]
	if !ok {
		return nil, nil, nil
	}
	key := f.byID[id]
	return key, f.owners[key.UserID], nil
}
func (f *fakeAPIKeyRepo) ListByUser(_ context.Context, userID string) ([]domain.APIKey, error) {
	out := []domain.APIKey{}
	for _, k := range f.byID {
		if k.UserID == userID {
			out = append(out, *k)
		}
	}
	return out, nil
}
func (f *fakeAPIKeyRepo) List(context.Context) ([]domain.APIKey, error) {
	out := []domain.APIKey{}
	for _, k := range f.byID {
		out = append(out, *k)
	}
	return out, nil
}
func (f *fakeAPIKeyRepo) Create(_ context.Context, k *domain.APIKey, hash, _ string) (string, error) {
	id := "key-" + k.Prefix
	stored := *k
	stored.ID = id
	f.byID[id] = &stored
	f.byHash[hash] = id
	return id, nil
}
func (f *fakeAPIKeyRepo) Revoke(_ context.Context, id, _ string) (bool, error) {
	k, ok := f.byID[id]
	if !ok || k.RevokedAt != nil {
		return false, nil
	}
	now := time.Now()
	k.RevokedAt = &now
	return true, nil
}
func (f *fakeAPIKeyRepo) TouchLastUsed(context.Context, string) error {
	f.touch++
	return nil
}

type fakeUsers struct{ byID map[string]*domain.User }

func (f *fakeUsers) FindByID(_ context.Context, id string) (*domain.User, error) {
	return f.byID[id], nil
}
func (f *fakeUsers) FindByEmail(context.Context, string) (*domain.User, string, error) {
	return nil, "", nil
}
func (f *fakeUsers) Create(context.Context, string, string, string, domain.Role) (string, error) {
	return "", nil
}
func (f *fakeUsers) UpdatePassword(context.Context, string, string) error        { return nil }
func (f *fakeUsers) UpdateRole(context.Context, string, domain.Role) error       { return nil }
func (f *fakeUsers) UpdateProfile(context.Context, string, string, string) error { return nil }

func keyFixture(t *testing.T) (*APIKeys, *fakeAPIKeyRepo, *domain.User, *domain.User) {
	t.Helper()

	admin := &domain.User{ID: "user-admin", FullName: "Demo Admin", Email: "admin@x.id", Role: domain.RoleAdmin}
	member := &domain.User{ID: "user-member", FullName: "Demo Member", Email: "member@x.id", Role: domain.RoleMember}

	repo := newFakeAPIKeyRepo()
	repo.owners[admin.ID] = admin
	repo.owners[member.ID] = member

	users := &fakeUsers{byID: map[string]*domain.User{admin.ID: admin, member.ID: member}}
	return NewAPIKeys(repo, users), repo, admin, member
}

// The secret is returned once. Everything after that is the digest, so a key
// that leaks from the database is not a key that opens anything.
func TestCreateReturnsTheKeyOnceAndStoresOnlyItsDigest(t *testing.T) {
	uc, repo, admin, _ := keyFixture(t)

	created, err := uc.Create(context.Background(), CreateAPIKeyInput{Nama: "Integrasi"}, admin)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if !strings.HasPrefix(created.Key, KeyPrefix) {
		t.Errorf("the key is not recognisable as ours: %q", created.Key)
	}
	if created.Record.Prefix != created.Key[:visiblePrefix] {
		t.Errorf("the stored prefix does not match the key")
	}

	// Nothing anywhere in the stored record repeats the secret.
	for _, stored := range repo.byID {
		if strings.Contains(created.Key, stored.Prefix) && len(stored.Prefix) > visiblePrefix {
			t.Error("more of the key is stored than the visible prefix")
		}
	}
	for hash := range repo.byHash {
		if hash == created.Key {
			t.Fatal("the key itself is stored, not a digest")
		}
	}
}

func TestCreatedKeysAreDistinct(t *testing.T) {
	uc, _, admin, _ := keyFixture(t)

	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		created, err := uc.Create(context.Background(), CreateAPIKeyInput{Nama: "k"}, admin)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if seen[created.Key] {
			t.Fatal("two keys came out the same")
		}
		seen[created.Key] = true
	}
}

func TestAuthenticateAcceptsALiveKey(t *testing.T) {
	uc, _, admin, _ := keyFixture(t)

	created, err := uc.Create(context.Background(), CreateAPIKeyInput{Nama: "Integrasi"}, admin)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	user, key, err := uc.Authenticate(context.Background(), created.Key)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if user == nil || key == nil {
		t.Fatal("a fresh key was rejected")
	}
	// The key acts as its owner, which is what makes the existing role
	// guards apply unchanged.
	if user.ID != admin.ID || user.Role != domain.RoleAdmin {
		t.Errorf("authenticated as %+v", user)
	}
}

// Unknown, revoked and expired all answer the same way, so a caller cannot
// use the difference to learn which keys exist.
func TestAuthenticateRejectsAnythingUnusable(t *testing.T) {
	uc, repo, admin, _ := keyFixture(t)
	ctx := context.Background()

	live, _ := uc.Create(ctx, CreateAPIKeyInput{Nama: "hidup"}, admin)

	revoked, _ := uc.Create(ctx, CreateAPIKeyInput{Nama: "dicabut"}, admin)
	if err := uc.Revoke(ctx, revoked.Record.ID, admin); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	expired, _ := uc.Create(ctx, CreateAPIKeyInput{Nama: "kedaluwarsa", ExpiresInDays: 1}, admin)
	past := time.Now().AddDate(0, 0, -1)
	repo.byID[expired.Record.ID].ExpiresAt = &past

	cases := []struct {
		name      string
		presented string
		want      bool
	}{
		{"live", live.Key, true},
		{"revoked", revoked.Key, false},
		{"expired", expired.Key, false},
		{"unknown but well-formed", KeyPrefix + "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", false},
		{"not one of ours", "sk-something-else", false},
		{"empty", "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			user, _, err := uc.Authenticate(ctx, c.presented)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (user != nil) != c.want {
				t.Errorf("accepted = %v, want %v", user != nil, c.want)
			}
		})
	}
}

// Surrounding whitespace is what a copied-and-pasted key arrives with.
func TestAuthenticateToleratesPaddingAroundTheKey(t *testing.T) {
	uc, _, admin, _ := keyFixture(t)

	created, _ := uc.Create(context.Background(), CreateAPIKeyInput{Nama: "k"}, admin)
	user, _, err := uc.Authenticate(context.Background(), "  "+created.Key+"\n")
	if err != nil || user == nil {
		t.Errorf("a padded key was rejected: user=%v err=%v", user, err)
	}
}

// A key hands out its owner's access, so issuing one for somebody else is
// itself an administrative act.
func TestOnlyAdminsIssueKeysForOtherPeople(t *testing.T) {
	uc, _, admin, member := keyFixture(t)
	ctx := context.Background()

	_, err := uc.Create(ctx, CreateAPIKeyInput{Nama: "k", UserID: admin.ID}, member)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("a member issuing for an admin got %v, want forbidden", err)
	}

	if _, err := uc.Create(ctx, CreateAPIKeyInput{Nama: "k", UserID: member.ID}, admin); err != nil {
		t.Errorf("an admin issuing for a member got %v", err)
	}
	if _, err := uc.Create(ctx, CreateAPIKeyInput{Nama: "k"}, member); err != nil {
		t.Errorf("a member issuing for themselves got %v", err)
	}
}

func TestCreateValidatesItsInput(t *testing.T) {
	uc, _, admin, _ := keyFixture(t)

	cases := []struct {
		name string
		in   CreateAPIKeyInput
	}{
		{"no name", CreateAPIKeyInput{}},
		{"blank name", CreateAPIKeyInput{Nama: "   "}},
		{"negative expiry", CreateAPIKeyInput{Nama: "k", ExpiresInDays: -1}},
		{"absurd expiry", CreateAPIKeyInput{Nama: "k", ExpiresInDays: 99999}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := uc.Create(context.Background(), c.in, admin); !errors.Is(err, domain.ErrInvalidInput) {
				t.Errorf("got %v, want an invalid-input error", err)
			}
		})
	}
}

func TestRevokeIsImmediateAndOnlyOnce(t *testing.T) {
	uc, _, admin, _ := keyFixture(t)
	ctx := context.Background()

	created, _ := uc.Create(ctx, CreateAPIKeyInput{Nama: "k"}, admin)

	if err := uc.Revoke(ctx, created.Record.ID, admin); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if user, _, _ := uc.Authenticate(ctx, created.Key); user != nil {
		t.Error("a revoked key still authenticates")
	}
	if err := uc.Revoke(ctx, created.Record.ID, admin); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("a second revoke gave %v, want a conflict", err)
	}
}

// One member must not be able to revoke another's key, and the refusal must
// not confirm that the key exists.
func TestMembersCannotTouchSomeoneElsesKey(t *testing.T) {
	uc, _, admin, member := keyFixture(t)
	ctx := context.Background()

	created, _ := uc.Create(ctx, CreateAPIKeyInput{Nama: "milik admin"}, admin)

	if err := uc.Revoke(ctx, created.Record.ID, member); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("got %v, want not-found", err)
	}
	if user, _, _ := uc.Authenticate(ctx, created.Key); user == nil {
		t.Error("the key stopped working despite the refusal")
	}
}

func TestListIsScopedToTheCaller(t *testing.T) {
	uc, _, admin, member := keyFixture(t)
	ctx := context.Background()

	if _, err := uc.Create(ctx, CreateAPIKeyInput{Nama: "a"}, admin); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Create(ctx, CreateAPIKeyInput{Nama: "b"}, member); err != nil {
		t.Fatal(err)
	}

	adminSees, _ := uc.List(ctx, admin)
	if len(adminSees) != 2 {
		t.Errorf("an admin sees %d keys, want every one", len(adminSees))
	}

	memberSees, _ := uc.List(ctx, member)
	if len(memberSees) != 1 {
		t.Fatalf("a member sees %d keys, want only their own", len(memberSees))
	}
	if memberSees[0].UserID != member.ID {
		t.Error("a member can see somebody else's key")
	}
}

func TestReadOnlyIsRecorded(t *testing.T) {
	uc, _, admin, _ := keyFixture(t)
	ctx := context.Background()

	read, _ := uc.Create(ctx, CreateAPIKeyInput{Nama: "baca", ReadOnly: true}, admin)
	write, _ := uc.Create(ctx, CreateAPIKeyInput{Nama: "tulis", ReadOnly: false}, admin)

	if !read.Record.ReadOnly {
		t.Error("the read-only key is not marked read-only")
	}
	if write.Record.ReadOnly {
		t.Error("the write key is marked read-only")
	}
}
