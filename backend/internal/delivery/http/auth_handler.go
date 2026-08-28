package http

import (
	"net/http"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(usecase.SessionDuration),
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	user, token, err := s.auth.SignIn(r.Context(), body.Email, body.Password)
	if err != nil {
		fail(w, err)
		return
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(s.cfg.CookieName); err == nil {
		if err := s.auth.SignOut(r.Context(), cookie.Value); err != nil {
			fail(w, err)
			return
		}
	}
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleMe returns the caller plus their competition profile, which is what
// the frontend needs to decide what to render on boot.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	member, err := s.members.Profile(r.Context(), user.ID)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user, "member": member})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	user := userFrom(r.Context())
	if err := s.auth.ChangePassword(r.Context(), user.ID, body.CurrentPassword, body.NewPassword); err != nil {
		fail(w, err)
		return
	}

	// Every session was dropped, including this one.
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"ok":        true,
		"service":   "brag-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.db.Ping(r.Context()); err != nil {
		status["ok"] = false
		status["database"] = "unreachable"
		writeJSON(w, http.StatusServiceUnavailable, status)
		return
	}

	status["database"] = "ok"
	writeJSON(w, http.StatusOK, status)
}
