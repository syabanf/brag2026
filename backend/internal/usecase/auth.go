// Package usecase holds application rules. It orchestrates domain entities and
// repositories, and knows nothing about HTTP or SQL.
package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

const SessionDuration = 30 * 24 * time.Hour

// dummyHash is a valid bcrypt digest of a value nobody can supply. Comparing
// against it makes a missing account cost the same as a wrong password.
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

type Auth struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
}

func NewAuth(users domain.UserRepository, sessions domain.SessionRepository) *Auth {
	return &Auth{users: users, sessions: sessions}
}

// HashToken is exported so the HTTP layer can look a cookie up without
// learning how tokens are stored.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// SignIn verifies the password and issues a session token. The error is
// deliberately identical for unknown emails and wrong passwords so it cannot
// be used to enumerate accounts.
func (a *Auth) SignIn(ctx context.Context, email, password string) (*domain.User, string, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return nil, "", domain.Invalid("Email dan kata sandi wajib diisi.")
	}

	user, hash, err := a.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}

	// An unknown email still pays the cost of a bcrypt comparison. Returning
	// early would answer in about a millisecond instead of sixty, which is a
	// reliable oracle for enumerating who has an account.
	if user == nil {
		hash = dummyHash
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil || user == nil {
		return nil, "", domain.NewError(domain.ErrUnauthorized, "Email atau kata sandi salah.")
	}

	token, err := newToken()
	if err != nil {
		return nil, "", err
	}

	if err := a.sessions.Create(ctx, user.ID, HashToken(token), time.Now().Add(SessionDuration)); err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (a *Auth) UserForToken(ctx context.Context, token string) (*domain.User, error) {
	if token == "" {
		return nil, nil
	}
	return a.sessions.FindUserByTokenHash(ctx, HashToken(token))
}

func (a *Auth) SignOut(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return a.sessions.Delete(ctx, HashToken(token))
}

// ChangePassword requires the current password, so a stolen session alone
// cannot lock the real owner out.
func (a *Auth) ChangePassword(ctx context.Context, userID, current, next string) error {
	if len(next) < 6 {
		return domain.Invalid("Kata sandi baru minimal 6 karakter.")
	}

	user, err := a.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.NotFound("Pengguna tidak ditemukan.")
	}

	_, hash, err := a.users.FindByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(current)) != nil {
		return domain.Invalid("Kata sandi saat ini salah.")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(next), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := a.users.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return err
	}

	// Other sessions are dropped so a compromised one dies with the change.
	return a.sessions.DeleteAllForUser(ctx, userID)
}

// SetPassword is the administrative reset: no current password required.
func (a *Auth) SetPassword(ctx context.Context, userID, next string) error {
	if len(next) < 6 {
		return domain.Invalid("Kata sandi minimal 6 karakter.")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(next), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return a.users.UpdatePassword(ctx, userID, string(hash))
}

func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(hash), err
}
