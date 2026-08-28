package http

import (
	"context"
	"net/http"

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
func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
