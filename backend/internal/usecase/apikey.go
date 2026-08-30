package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// KeyPrefix marks a BRAG key on sight, so one found in a log or a config file
// is recognisable without being tried against the API.
const KeyPrefix = "brag_"

// keyBytes is the entropy behind each key. 32 bytes is past the point where
// guessing is a consideration, which is why the digest below can be a plain
// SHA-256 rather than a deliberately slow hash.
const keyBytes = 32

// visiblePrefix is how much of the key is kept in clear for identification.
// Long enough to tell two keys apart in a list, far too short to be useful to
// anyone who reads it.
const visiblePrefix = 12

type APIKeys struct {
	keys  domain.APIKeyRepository
	users domain.UserRepository
}

func NewAPIKeys(keys domain.APIKeyRepository, users domain.UserRepository) *APIKeys {
	return &APIKeys{keys: keys, users: users}
}

// hashKey is the one place a key becomes a digest, so the storage and the
// lookup can never drift apart.
func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

type CreateAPIKeyInput struct {
	Nama string
	// UserID is who the key acts as. Empty means the caller themselves.
	UserID   string
	ReadOnly bool
	// ExpiresInDays of zero means the key does not expire on its own.
	ExpiresInDays int
}

// CreatedAPIKey carries the one and only sight of the secret. It is not stored
// anywhere in this form, so a caller that loses it has to issue another.
type CreatedAPIKey struct {
	Key    string         `json:"key"`
	Record *domain.APIKey `json:"record"`
}

func (u *APIKeys) Create(ctx context.Context, in CreateAPIKeyInput, actor *domain.User) (*CreatedAPIKey, error) {
	in.Nama = strings.TrimSpace(in.Nama)
	if in.Nama == "" {
		return nil, domain.Invalid("Nama kunci wajib diisi.")
	}
	if in.ExpiresInDays < 0 || in.ExpiresInDays > 3650 {
		return nil, domain.Invalid("Masa berlaku harus antara 0 dan 3650 hari.")
	}

	ownerID := in.UserID
	if ownerID == "" {
		ownerID = actor.ID
	}

	owner, err := u.users.FindByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if owner == nil {
		return nil, domain.NotFound("Pengguna tidak ditemukan.")
	}
	// A key inherits its owner's role, so issuing one for somebody else is a
	// way to hand out their access. Only an admin may do that, and even then
	// never above their own level.
	if owner.ID != actor.ID && !actor.Role.IsAdmin() {
		return nil, domain.Forbidden("Hanya admin yang bisa membuat kunci atas nama orang lain.")
	}
	if owner.Role.IsAdmin() && !actor.Role.IsAdmin() {
		return nil, domain.Forbidden("Tidak bisa membuat kunci dengan akses lebih tinggi dari milik Anda.")
	}

	raw := make([]byte, keyBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	key := KeyPrefix + base64.RawURLEncoding.EncodeToString(raw)

	record := &domain.APIKey{
		Nama:     in.Nama,
		Prefix:   key[:visiblePrefix],
		UserID:   owner.ID,
		ReadOnly: in.ReadOnly,
	}
	if in.ExpiresInDays > 0 {
		expiry := time.Now().AddDate(0, 0, in.ExpiresInDays)
		record.ExpiresAt = &expiry
	}

	id, err := u.keys.Create(ctx, record, hashKey(key), actor.ID)
	if err != nil {
		return nil, err
	}

	record.ID = id
	record.CreatedAt = time.Now()
	record.UserName, record.UserEmail = owner.FullName, owner.Email

	return &CreatedAPIKey{Key: key, Record: record}, nil
}

// Authenticate resolves a presented key to the user it acts as. It returns nil
// without an error for anything unusable — unknown, revoked or expired — so a
// caller cannot tell those apart by the response.
func (u *APIKeys) Authenticate(ctx context.Context, presented string) (*domain.User, *domain.APIKey, error) {
	presented = strings.TrimSpace(presented)
	if !strings.HasPrefix(presented, KeyPrefix) {
		return nil, nil, nil
	}

	key, user, err := u.keys.FindByHash(ctx, hashKey(presented))
	if err != nil {
		return nil, nil, err
	}
	if key == nil || user == nil || !key.Active(time.Now()) {
		return nil, nil, nil
	}

	return user, key, nil
}

func (u *APIKeys) List(ctx context.Context, actor *domain.User) ([]domain.APIKey, error) {
	// An admin manages every key; anyone else only ever sees their own.
	if actor.Role.IsAdmin() {
		return u.keys.List(ctx)
	}
	return u.keys.ListByUser(ctx, actor.ID)
}

func (u *APIKeys) Revoke(ctx context.Context, id string, actor *domain.User) error {
	// Finding the key first is what makes the ownership check possible, and
	// what lets a second press say "already revoked" rather than "not found".
	keys, err := u.List(ctx, actor)
	if err != nil {
		return err
	}

	var found *domain.APIKey
	for i := range keys {
		if keys[i].ID == id {
			found = &keys[i]
			break
		}
	}
	if found == nil {
		return domain.NotFound("Kunci tidak ditemukan.")
	}
	if found.RevokedAt != nil {
		return domain.Conflict("Kunci ini sudah dicabut.")
	}

	ok, err := u.keys.Revoke(ctx, id, actor.ID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.Conflict("Kunci ini sudah dicabut.")
	}
	return nil
}

// TouchQuietly records that a key was used. It is deliberately silent: the
// request it authenticated has already been authorised, and failing it over a
// bookkeeping write would be the wrong trade.
func (u *APIKeys) TouchQuietly(ctx context.Context, keyID string) {
	_ = u.keys.TouchLastUsed(ctx, keyID)
}
