package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
	"github.com/ikurniawann/brag2026/backend/internal/usecase"
)

// dayParam reads an optional ?day=YYYY-MM-DD, defaulting to today so the
// common case needs no argument.
func dayParam(r *http.Request) (time.Time, error) {
	raw := r.URL.Query().Get("day")
	if raw == "" {
		return time.Now(), nil
	}
	return parseDate(raw)
}

// ── Weekly events ─────────────────────────────────────────────────────────

func (s *Server) handleListWeeklyEvents(w http.ResponseWriter, r *http.Request) {
	season, err := s.seasons.FindActive(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if season == nil {
		writeJSON(w, http.StatusOK, []domain.WeeklyEvent{})
		return
	}

	events, err := s.events.ListBySeason(r.Context(), season.ID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// handleEventBank exposes the twelve codes with their copy, so the admin
// picker does not hardcode Indonesian strings in the frontend.
func (s *Server) handleEventBank(w http.ResponseWriter, r *http.Request) {
	type entry struct {
		Code    string `json:"code"`
		Nama    string `json:"nama"`
		Mekanik string `json:"mekanik"`
	}

	out := []entry{}
	for _, code := range domain.EventCodes() {
		nama, mekanik := domain.DescribeEvent(code)
		out = append(out, entry{Code: string(code), Nama: nama, Mekanik: mekanik})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleScheduleEvent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MingguKe               int     `json:"minggu_ke"`
		EventCode              string  `json:"event_code"`
		TargetClassificationID *string `json:"target_classification_id"`
		TanggalMulai           string  `json:"tanggal_mulai"`
		TanggalSelesai         string  `json:"tanggal_selesai"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	if body.MingguKe < 1 || body.MingguKe > 12 {
		fail(w, domain.Invalid("Minggu harus antara 1 dan 12."))
		return
	}
	if !domain.ValidEventCode(body.EventCode) {
		fail(w, domain.Invalid("Kode event tidak dikenal."))
		return
	}

	mulai, err := parseDate(body.TanggalMulai)
	if err != nil {
		fail(w, err)
		return
	}
	selesai, err := parseDate(body.TanggalSelesai)
	if err != nil {
		fail(w, err)
		return
	}
	if selesai.Before(mulai) {
		fail(w, domain.Invalid("Tanggal selesai tidak boleh sebelum tanggal mulai."))
		return
	}

	season, err := s.seasons.FindActive(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if season == nil {
		fail(w, domain.NotFound("Season aktif tidak ditemukan."))
		return
	}

	id, err := s.events.Upsert(r.Context(), &domain.WeeklyEvent{
		SeasonID:               season.ID,
		MingguKe:               body.MingguKe,
		EventCode:              domain.EventCode(body.EventCode),
		TargetClassificationID: body.TargetClassificationID,
		TanggalMulai:           mulai,
		TanggalSelesai:         selesai,
	})
	if err != nil {
		fail(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

func (s *Server) handleDeleteWeeklyEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.events.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── Scoring passes ────────────────────────────────────────────────────────

func (s *Server) handleWeeklyPass(w http.ResponseWriter, r *http.Request) {
	day, err := dayParam(r)
	if err != nil {
		fail(w, err)
		return
	}

	result, err := s.passes.RunWeekly(r.Context(), day)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDailyPass(w http.ResponseWriter, r *http.Request) {
	day, err := dayParam(r)
	if err != nil {
		fail(w, err)
		return
	}

	result, err := s.passes.RunDaily(r.Context(), day)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ── Prize pool ────────────────────────────────────────────────────────────

func (s *Server) handleListPrizes(w http.ResponseWriter, r *http.Request) {
	prizes, err := s.prizes.List(r.Context(), r.URL.Query().Get("status"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, prizes)
}

type prizeBody struct {
	NamaHadiah     string   `json:"nama_hadiah"`
	Deskripsi      *string  `json:"deskripsi"`
	NilaiEstimasi  *float64 `json:"nilai_estimasi"`
	Alokasi        string   `json:"alokasi"`
	KategoriTarget *string  `json:"kategori_target"`
}

func (b prizeBody) toInput() usecase.PrizeInput {
	return usecase.PrizeInput{
		NamaHadiah:     b.NamaHadiah,
		Deskripsi:      b.Deskripsi,
		NilaiEstimasi:  b.NilaiEstimasi,
		Alokasi:        b.Alokasi,
		KategoriTarget: b.KategoriTarget,
	}
}

// handleDonatePrize is the member path: the donation waits for approval.
func (s *Server) handleDonatePrize(w http.ResponseWriter, r *http.Request) {
	var body prizeBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	member := memberFrom(r.Context())
	id, err := s.prizes.Donate(r.Context(), body.toInput(), member.ID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// handleSeedPrize is the committee path: approved on arrival.
func (s *Server) handleSeedPrize(w http.ResponseWriter, r *http.Request) {
	var body prizeBody
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	id, err := s.prizes.Seed(r.Context(), body.toInput())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleSetPrizeStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status     string  `json:"status"`
		PemenangID *string `json:"pemenang_id"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, err)
		return
	}

	if err := s.prizes.SetStatus(r.Context(), chi.URLParam(r, "id"), body.Status, body.PemenangID); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeletePrize(w http.ResponseWriter, r *http.Request) {
	if err := s.prizes.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleTicketStandings(w http.ResponseWriter, r *http.Request) {
	tickets, err := s.prizes.TicketStandings(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (s *Server) handleIssueTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := s.prizes.IssueTickets(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

// handleDrawPrize runs the lottery for one prize and returns the winner, so
// the admin screen can show the name without a second round trip.
func (s *Server) handleDrawPrize(w http.ResponseWriter, r *http.Request) {
	prize, err := s.prizes.Draw(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, prize)
}
