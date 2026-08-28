package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

const dateLayout = "2006-01-02"

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, domain.Invalid("Tanggal wajib diisi.")
	}
	// Accept both a plain date and a full timestamp, since browsers send either.
	if t, err := time.Parse(dateLayout, value); err == nil {
		return t, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, domain.Invalid("Format tanggal tidak valid.")
	}
	return t, nil
}

// ── TYFCB ─────────────────────────────────────────────────────────────────

type tyfcbSubmitBody struct {
	BuyerID string  `json:"buyer_id"`
	Nilai   float64 `json:"nilai"`
	Tanggal string  `json:"tanggal"`
	// MemberID lets a captain file on behalf of a team member.
	MemberID string `json:"member_id,omitempty"`
}

func (s *Server) handleSubmitTyfcb(w http.ResponseWriter, r *http.Request) {
	var body tyfcbSubmitBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	tanggal, err := parseDate(body.Tanggal)
	if err != nil {
		fail(w, err)
		return
	}

	user := userFrom(r.Context())
	member := memberFrom(r.Context())

	entry, err := s.tyfcb.Submit(r.Context(), usecase.SubmitTyfcbInput{
		BuyerID:  body.BuyerID,
		SellerID: member.ID,
		Nilai:    body.Nilai,
		Tanggal:  tanggal,
	}, &user.ID)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

// handleCaptainSubmitTyfcb records a transaction for someone else on the
// captain's team; the seller is the named member, not the captain.
func (s *Server) handleCaptainSubmitTyfcb(w http.ResponseWriter, r *http.Request) {
	var body tyfcbSubmitBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}
	if body.MemberID == "" {
		fail(w, domain.Invalid("member_id wajib diisi."))
		return
	}

	tanggal, err := parseDate(body.Tanggal)
	if err != nil {
		fail(w, err)
		return
	}

	user := userFrom(r.Context())
	if err := s.assertSameTeam(r, body.MemberID); err != nil {
		fail(w, err)
		return
	}

	entry, err := s.tyfcb.Submit(r.Context(), usecase.SubmitTyfcbInput{
		BuyerID:  body.BuyerID,
		SellerID: body.MemberID,
		Nilai:    body.Nilai,
		Tanggal:  tanggal,
	}, &user.ID)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

func (s *Server) handleListTyfcb(w http.ResponseWriter, r *http.Request) {
	season, err := s.seasons.FindActive(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if season == nil {
		writeJSON(w, http.StatusOK, []domain.TyfcbEntry{})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	entries, err := s.tyfcb.List(r.Context(), domain.TyfcbFilter{
		SeasonID: season.ID,
		Status:   r.URL.Query().Get("status"),
		TeamID:   r.URL.Query().Get("team_id"),
		Limit:    limit,
	})
	if err != nil {
		fail(w, err)
		return
	}

	counts, err := s.tyfcb.CountByStatus(r.Context(), season.ID)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "counts": counts})
}

func (s *Server) handleSetTyfcbStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	user := userFrom(r.Context())
	id := chi.URLParam(r, "id")

	if err := s.tyfcb.SetStatus(r.Context(), id, domain.TyfcbStatus(body.Status), user.ID); err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": body.Status})
}

func (s *Server) handleVoidTyfcb(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	if err := s.tyfcb.Void(r.Context(), chi.URLParam(r, "id"), user.ID); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Visitors ──────────────────────────────────────────────────────────────

type visitorBody struct {
	Nama          string `json:"nama"`
	Kontak        string `json:"kontak"`
	TanggalUndang string `json:"tanggal_undang"`
	MemberID      string `json:"member_id,omitempty"`
}

func (s *Server) handleRegisterVisitor(w http.ResponseWriter, r *http.Request) {
	var body visitorBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	tanggal, err := parseDate(body.TanggalUndang)
	if err != nil {
		fail(w, err)
		return
	}

	user := userFrom(r.Context())
	member := memberFrom(r.Context())

	visitor, err := s.visitors.Register(r.Context(), usecase.RegisterVisitorInput{
		Nama:          body.Nama,
		Kontak:        body.Kontak,
		TanggalUndang: tanggal,
		InviterID:     member.ID,
	}, &user.ID)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, visitor)
}

func (s *Server) handleCaptainRegisterVisitor(w http.ResponseWriter, r *http.Request) {
	var body visitorBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}
	if body.MemberID == "" {
		fail(w, domain.Invalid("member_id wajib diisi."))
		return
	}

	tanggal, err := parseDate(body.TanggalUndang)
	if err != nil {
		fail(w, err)
		return
	}

	if err := s.assertSameTeam(r, body.MemberID); err != nil {
		fail(w, err)
		return
	}

	user := userFrom(r.Context())
	visitor, err := s.visitors.Register(r.Context(), usecase.RegisterVisitorInput{
		Nama:          body.Nama,
		Kontak:        body.Kontak,
		TanggalUndang: tanggal,
		InviterID:     body.MemberID,
	}, &user.ID)
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, visitor)
}

func (s *Server) handleListVisitors(w http.ResponseWriter, r *http.Request) {
	season, err := s.seasons.FindActive(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if season == nil {
		writeJSON(w, http.StatusOK, []domain.Visitor{})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	visitors, err := s.visitors.List(r.Context(), domain.VisitorFilter{
		SeasonID:  season.ID,
		Status:    r.URL.Query().Get("status"),
		TeamID:    r.URL.Query().Get("team_id"),
		InviterID: r.URL.Query().Get("inviter_id"),
		Limit:     limit,
	})
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, visitors)
}

func (s *Server) handleUpdateVisitor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StatusHadir *string `json:"status_hadir"`
		IsConverted *bool   `json:"is_converted"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	err := s.visitors.Update(r.Context(), chi.URLParam(r, "id"), usecase.UpdateVisitorInput{
		StatusHadir: body.StatusHadir,
		IsConverted: body.IsConverted,
	})
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleVoidVisitor(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	if err := s.visitors.Void(r.Context(), chi.URLParam(r, "id"), user.ID); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// assertSameTeam stops a captain from acting on members of another team.
// Admins bypass the check.
func (s *Server) assertSameTeam(r *http.Request, memberID string) error {
	user := userFrom(r.Context())
	if user.Role.IsAdmin() {
		return nil
	}

	actor, err := s.members.Profile(r.Context(), user.ID)
	if err != nil {
		return err
	}
	if actor == nil || actor.TeamID == nil {
		return domain.Forbidden("Anda belum tergabung dalam tim.")
	}

	target, err := s.memberRepo.FindByID(r.Context(), memberID)
	if err != nil {
		return err
	}
	if target == nil {
		return domain.NotFound("Member tidak ditemukan.")
	}
	if target.TeamID == nil || *target.TeamID != *actor.TeamID {
		return domain.Forbidden("Member ini bukan anggota tim Anda.")
	}
	return nil
}
