package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.apiKeys.List(r.Context(), userFrom(r.Context()))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, keys)
}

// handleCreateAPIKey returns the secret in its response and nowhere else. The
// client has to show it to the person there and then, because no later request
// can produce it again.
func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nama          string `json:"nama"`
		UserID        string `json:"user_id"`
		ReadOnly      *bool  `json:"read_only"`
		ExpiresInDays int    `json:"expires_in_days"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	// Read-only unless someone deliberately says otherwise: a key that can
	// write is the one worth thinking twice about.
	readOnly := true
	if body.ReadOnly != nil {
		readOnly = *body.ReadOnly
	}

	created, err := s.apiKeys.Create(r.Context(), usecase.CreateAPIKeyInput{
		Nama:          body.Nama,
		UserID:        body.UserID,
		ReadOnly:      readOnly,
		ExpiresInDays: body.ExpiresInDays,
	}, userFrom(r.Context()))
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	if err := s.apiKeys.Revoke(r.Context(), chi.URLParam(r, "id"), userFrom(r.Context())); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
