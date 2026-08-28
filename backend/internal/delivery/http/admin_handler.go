package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

// ── Members ───────────────────────────────────────────────────────────────

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.members.List(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (s *Server) handleSearchMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.members.Search(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (s *Server) handleCreateMember(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FullName      string  `json:"full_name"`
		Email         string  `json:"email"`
		Password      string  `json:"password"`
		TeamID        *string `json:"team_id"`
		KlasifikasiID *string `json:"klasifikasi_id"`
		ColorStatus   string  `json:"color_status"`
		Role          string  `json:"role"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	id, err := s.members.Create(r.Context(), usecase.CreateMemberInput{
		FullName:      body.FullName,
		Email:         body.Email,
		Password:      body.Password,
		TeamID:        body.TeamID,
		KlasifikasiID: body.KlasifikasiID,
		ColorStatus:   body.ColorStatus,
		Role:          body.Role,
	})
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleUpdateMember(w http.ResponseWriter, r *http.Request) {
	// Pointers distinguish "field omitted" from "field set to null", which the
	// edit form relies on to clear a team or classification.
	var body struct {
		FullName      *string `json:"full_name"`
		Email         *string `json:"email"`
		NewPassword   *string `json:"new_password"`
		TeamID        *string `json:"team_id"`
		KlasifikasiID *string `json:"klasifikasi_id"`
		ColorStatus   *string `json:"color_status"`
		IsActive      *bool   `json:"is_active"`
		Role          *string `json:"role"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	in := usecase.UpdateMemberInput{
		FullName:    body.FullName,
		Email:       body.Email,
		NewPassword: body.NewPassword,
		ColorStatus: body.ColorStatus,
		IsActive:    body.IsActive,
		Role:        body.Role,
	}
	if body.TeamID != nil {
		if *body.TeamID == "" {
			in.ClearTeam = true
		} else {
			in.TeamID = body.TeamID
		}
	}
	if body.KlasifikasiID != nil {
		if *body.KlasifikasiID == "" {
			in.ClearKlas = true
		} else {
			in.KlasifikasiID = body.KlasifikasiID
		}
	}

	if err := s.members.Update(r.Context(), chi.URLParam(r, "id"), in, s.auth); err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Teams ─────────────────────────────────────────────────────────────────

func (s *Server) handleListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := s.catalog.ListTeams(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, teams)
}

// handleTeamsMeta backs the member form: teams and classifications in one trip.
func (s *Server) handleTeamsMeta(w http.ResponseWriter, r *http.Request) {
	teams, err := s.catalog.ListTeams(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	classes, err := s.catalog.ListClassifications(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"teams": teams, "classifications": classes})
}

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamaTim string `json:"nama_tim"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	id, err := s.catalog.CreateTeam(r.Context(), body.NamaTim)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamaTim string `json:"nama_tim"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}
	if err := s.catalog.RenameTeam(r.Context(), chi.URLParam(r, "id"), body.NamaTim); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	if err := s.catalog.DeleteTeam(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Classifications ───────────────────────────────────────────────────────

func (s *Server) handleListClassifications(w http.ResponseWriter, r *http.Request) {
	classes, err := s.catalog.ListClassifications(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, classes)
}

func (s *Server) handleCreateClassification(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nama string `json:"nama"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	id, err := s.catalog.CreateClassification(r.Context(), body.Nama)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleUpdateClassification(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nama string `json:"nama"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}
	if err := s.catalog.RenameClassification(r.Context(), chi.URLParam(r, "id"), body.Nama); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteClassification(w http.ResponseWriter, r *http.Request) {
	if err := s.catalog.DeleteClassification(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Boosters ──────────────────────────────────────────────────────────────

func (s *Server) handleListBoosters(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"

	boosters, err := s.catalog.ListBoosters(r.Context(), activeOnly)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, boosters)
}

func (s *Server) handleGetBooster(w http.ResponseWriter, r *http.Request) {
	booster, err := s.catalog.FindBooster(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, booster)
}

type boosterBody struct {
	Judul           string  `json:"judul"`
	Deskripsi       *string `json:"deskripsi"`
	TanggalMulai    string  `json:"tanggal_mulai"`
	TanggalBerakhir string  `json:"tanggal_berakhir"`
	Poin            int     `json:"poin"`
	Status          string  `json:"status"`
}

func (b boosterBody) toInput() (usecase.BoosterInput, error) {
	mulai, err := parseDate(b.TanggalMulai)
	if err != nil {
		return usecase.BoosterInput{}, err
	}
	berakhir, err := parseDate(b.TanggalBerakhir)
	if err != nil {
		return usecase.BoosterInput{}, err
	}
	return usecase.BoosterInput{
		Judul:           b.Judul,
		Deskripsi:       b.Deskripsi,
		TanggalMulai:    mulai,
		TanggalBerakhir: berakhir,
		Poin:            b.Poin,
		Status:          b.Status,
	}, nil
}

func (s *Server) handleCreateBooster(w http.ResponseWriter, r *http.Request) {
	var body boosterBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}
	in, err := body.toInput()
	if err != nil {
		fail(w, err)
		return
	}

	id, err := s.catalog.CreateBooster(r.Context(), in)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleUpdateBooster(w http.ResponseWriter, r *http.Request) {
	var body boosterBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}
	in, err := body.toInput()
	if err != nil {
		fail(w, err)
		return
	}

	if err := s.catalog.UpdateBooster(r.Context(), chi.URLParam(r, "id"), in); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteBooster(w http.ResponseWriter, r *http.Request) {
	if err := s.catalog.DeleteBooster(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Leaderboard & dashboard ───────────────────────────────────────────────

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	data, err := s.leaderboard.Dashboard(r.Context(), user.ID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	teams, err := s.leaderboard.Standings(r.Context())
	if err != nil {
		fail(w, err)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	members, err := s.leaderboard.IndividualStandings(r.Context(), limit)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"teams": teams, "members": members})
}

func (s *Server) handleTeamHistory(w http.ResponseWriter, r *http.Request) {
	entries, err := s.leaderboard.TeamHistory(
		r.Context(), chi.URLParam(r, "id"), r.URL.Query().Get("kategori"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (s *Server) handleBadges(w http.ResponseWriter, r *http.Request) {
	badges, err := s.leaderboard.Badges(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, badges)
}

// ── Captain ───────────────────────────────────────────────────────────────

// handleCaptainTeam returns the captain's own team roster plus everything
// pending for it, which is the whole captain screen in one request.
func (s *Server) handleCaptainTeam(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	actor, err := s.members.Profile(r.Context(), user.ID)
	if err != nil {
		fail(w, err)
		return
	}
	if actor == nil || actor.TeamID == nil {
		writeError(w, http.StatusNotFound, "Anda belum tergabung dalam tim.")
		return
	}

	roster, err := s.members.ListByTeam(r.Context(), *actor.TeamID)
	if err != nil {
		fail(w, err)
		return
	}

	pending, err := s.tyfcb.List(r.Context(), domain.TyfcbFilter{
		SeasonID: actor.SeasonID,
		Status:   string(domain.TyfcbPending),
		TeamID:   *actor.TeamID,
	})
	if err != nil {
		fail(w, err)
		return
	}

	visitors, err := s.visitors.List(r.Context(), domain.VisitorFilter{
		SeasonID: actor.SeasonID,
		Status:   string(domain.VisitorTerdaftar),
		TeamID:   *actor.TeamID,
	})
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"team_id":            *actor.TeamID,
		"members":            roster,
		"pending_tyfcb":      pending,
		"terdaftar_visitors": visitors,
	})
}

func (s *Server) handleCaptainSetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NewPassword string `json:"new_password"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	user := userFrom(r.Context())

	// Admins may reset anyone; a captain is confined to their own team.
	var teamGuard *string
	if !user.Role.IsAdmin() {
		actor, err := s.members.Profile(r.Context(), user.ID)
		if err != nil {
			fail(w, err)
			return
		}
		if actor == nil || actor.TeamID == nil {
			writeError(w, http.StatusForbidden, "Anda belum tergabung dalam tim.")
			return
		}
		teamGuard = actor.TeamID
	}

	err := s.members.SetPasswordFor(r.Context(), teamGuard, chi.URLParam(r, "id"), body.NewPassword, s.auth)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
