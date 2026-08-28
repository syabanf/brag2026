package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

// ── Contact spheres ───────────────────────────────────────────────────────

func (s *Server) handleListSpheres(w http.ResponseWriter, r *http.Request) {
	spheres, err := s.network.ListSpheres(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, spheres)
}

func (s *Server) handleCreateSphere(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nama        string   `json:"nama"`
		Deskripsi   *string  `json:"deskripsi"`
		Klasifikasi []string `json:"klasifikasi_ids"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	id, err := s.network.CreateSphere(r.Context(), body.Nama, body.Deskripsi, body.Klasifikasi)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleSetSphereMembers(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Klasifikasi []string `json:"klasifikasi_ids"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	if err := s.network.SetSphereMembers(r.Context(), chi.URLParam(r, "id"), body.Klasifikasi); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteSphere(w http.ResponseWriter, r *http.Request) {
	if err := s.network.DeleteSphere(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── One-to-one logs ───────────────────────────────────────────────────────

// handleListOneToOne defaults to the caller's own meetings; admins can widen
// it to the whole season with ?all=true.
func (s *Server) handleListOneToOne(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	memberID := ""
	if r.URL.Query().Get("all") != "true" || !userFrom(r.Context()).Role.IsAdmin() {
		member, err := s.members.Profile(r.Context(), userFrom(r.Context()).ID)
		if err != nil {
			fail(w, err)
			return
		}
		if member == nil {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		memberID = member.ID
	}

	logs, err := s.network.ListOneToOne(r.Context(), memberID, limit)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (s *Server) handleLogOneToOne(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID string  `json:"member_id"`
		Tanggal  string  `json:"tanggal"`
		Catatan  *string `json:"catatan"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	tanggal, err := parseDate(body.Tanggal)
	if err != nil {
		fail(w, err)
		return
	}

	// The caller is always one side of their own meeting.
	user := userFrom(r.Context())
	member := memberFrom(r.Context())

	id, err := s.network.LogOneToOne(r.Context(), usecase.LogOneToOneInput{
		MemberA: member.ID,
		MemberB: body.MemberID,
		Tanggal: tanggal,
		Catatan: body.Catatan,
	}, &user.ID)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleDeleteOneToOne(w http.ResponseWriter, r *http.Request) {
	if err := s.network.DeleteOneToOne(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
