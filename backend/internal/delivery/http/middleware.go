package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type ctxKey string

const (
	ctxUser   ctxKey = "user"
	ctxMember ctxKey = "member"
)

func userFrom(ctx context.Context) *domain.User {
	u, _ := ctx.Value(ctxUser).(*domain.User)
	return u
}

func memberFrom(ctx context.Context) *domain.Member {
	m, _ := ctx.Value(ctxMember).(*domain.Member)
	return m
}

// authenticate resolves the session cookie onto a user, but does not reject
// anonymous callers — that is requireAuth's job. Splitting them lets public
// endpoints still personalise when a session happens to be present.
// presentedKey pulls an API key from either header. Authorization is what a
// client library will reach for; X-API-Key is what a curl example is easiest
// to write with, and both are common enough that supporting one would just
// mean answering the question repeatedly.
func presentedKey(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-API-Key")); v != "" {
		return v
	}
	if v := r.Header.Get("Authorization"); len(v) > 7 && strings.EqualFold(v[:7], "bearer ") {
		return strings.TrimSpace(v[7:])
	}
	return ""
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A key is checked first: a request that presents one is a machine,
		// and should not silently fall back to whatever cookie the browser
		// happened to attach.
		if presented := presentedKey(r); presented != "" {
			user, key, err := s.apiKeys.Authenticate(r.Context(), presented)
			if err != nil {
				fail(w, err)
				return
			}
			if user == nil {
				// Unknown, revoked and expired are all one answer, so a caller
				// cannot use the difference to probe.
				fail(w, domain.NewError(domain.ErrUnauthorized, "Kunci API tidak valid."))
				return
			}
			// Read-only is the only permission a key carries of its own; the
			// role checks downstream are the owner's, unchanged.
			if key.ReadOnly && r.Method != http.MethodGet && r.Method != http.MethodHead {
				fail(w, domain.Forbidden("Kunci API ini hanya untuk membaca."))
				return
			}

			s.apiKeys.TouchQuietly(r.Context(), key.ID)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUser, user)))
			return
		}

		cookie, err := r.Cookie(s.cfg.CookieName)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := s.auth.UserForToken(r.Context(), cookie.Value)
		if err != nil {
			fail(w, err)
			return
		}
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUser, user)))
	})
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userFrom(r.Context()) == nil {
			writeError(w, http.StatusUnauthorized, "Silakan masuk terlebih dahulu.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userFrom(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "Silakan masuk terlebih dahulu.")
			return
		}
		if !user.Role.IsAdmin() {
			writeError(w, http.StatusForbidden, "Hanya admin yang bisa mengakses ini.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireCaptain also admits admins, matching the original app where an admin
// can do anything a captain can.
func requireCaptain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userFrom(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "Silakan masuk terlebih dahulu.")
			return
		}
		if !user.Role.IsCaptain() {
			writeError(w, http.StatusForbidden, "Hanya kapten tim yang bisa mengakses ini.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withMember attaches the caller's competition profile. Endpoints that record
// contributions need it, and failing here beats a confusing null downstream.
func (s *Server) withMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userFrom(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "Silakan masuk terlebih dahulu.")
			return
		}

		member, err := s.members.Profile(r.Context(), user.ID)
		if err != nil {
			fail(w, err)
			return
		}
		if member == nil {
			writeError(w, http.StatusNotFound, "Profil member tidak ditemukan di season ini.")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxMember, member)))
	})
}
