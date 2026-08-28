package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// TourStep is one stop on the guided walkthrough. Copy lives here rather than
// in the frontend so the narration and the caption can never drift apart.
type TourStep struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Route string `json:"route"`
	// Voice is what gets synthesised; it reads more naturally aloud than Body.
	Voice string `json:"-"`
}

var tourSteps = []TourStep{
	{
		ID: "welcome", Route: "/",
		Title: "Selamat datang di BRAG 2026",
		Body:  "BRAG adalah platform gamifikasi untuk BNI Grow Annual Challenge. Satu musim berjalan 12 minggu, dan setiap kontribusi anggota menambah poin untuk timnya.",
		Voice: "Selamat datang di BRAG 2026. BRAG adalah platform gamifikasi untuk BNI Grow Annual Challenge. Satu musim berjalan dua belas minggu, dan setiap kontribusi anggota menambah poin untuk timnya.",
	},
	{
		ID: "dashboard", Route: "/",
		Title: "Dashboard anggota",
		Body:  "Di sini anggota melihat skor pribadi, posisi timnya, dan booster yang sedang aktif. Semua angka diambil dari score ledger — satu-satunya sumber kebenaran perhitungan poin.",
		Voice: "Ini dashboard anggota. Di sini anggota melihat skor pribadi, posisi timnya, dan booster yang sedang aktif. Semua angka diambil dari score ledger, satu-satunya sumber kebenaran perhitungan poin.",
	},
	{
		ID: "leaderboard", Route: "/leaderboard",
		Title: "Leaderboard tim",
		Body:  "Sepuluh tim diurutkan berdasarkan total poin. Tap salah satu tim untuk melihat rincian riwayat TYFCB dan visitor yang menyusun skornya.",
		Voice: "Ini leaderboard tim. Sepuluh tim diurutkan berdasarkan total poin. Tap salah satu tim untuk melihat rincian riwayat TYFCB dan visitor yang menyusun skornya.",
	},
	{
		ID: "submit", Route: "/submit",
		Title: "Catat kontribusi",
		Body:  "Anggota mencatat dua jenis kontribusi: TYFCB dan visitor. Setiap pengajuan masuk berstatus pending sampai diverifikasi admin.",
		Voice: "Di halaman ini anggota mencatat kontribusi. Ada dua jenis: TYFCB dan visitor. Setiap pengajuan masuk berstatus pending sampai diverifikasi admin.",
	},
	{
		ID: "scoring", Route: "/submit",
		Title: "Cara poin dihitung",
		Body:  "Poin TYFCB dihitung dari Band dikali Pair Penalty dikali Event Multiplier. Band naik bertingkat dari 10 poin untuk nilai di bawah 500 ribu, sampai 200 poin untuk 500 juta ke atas.",
		Voice: "Bagaimana poin dihitung? Poin TYFCB dihitung dari Band, dikali Pair Penalty, dikali Event Multiplier. Band naik bertingkat dari sepuluh poin untuk nilai di bawah lima ratus ribu, sampai dua ratus poin untuk lima ratus juta ke atas.",
	},
	{
		ID: "events", Route: "/booster",
		Title: "Event mingguan",
		Body:  "Setiap minggu punya satu event aktif yang mengubah pengali poin. Founder's Frenzy melipatkan semua skor 1,5 kali; Spread the Love menggandakan TYFCB ke pasangan baru.",
		Voice: "Setiap minggu punya satu event aktif yang mengubah pengali poin. Founder's Frenzy melipatkan semua skor satu setengah kali. Spread the Love menggandakan TYFCB ke pasangan baru.",
	},
	{
		ID: "badges", Route: "/awards",
		Title: "Badge otomatis",
		Body:  "Dua belas badge diberikan otomatis begitu anggota mencapai milestone — dari First Blood untuk TYFCB pertama, sampai Centurion saat skor pribadi melewati 100.",
		Voice: "Dua belas badge diberikan otomatis begitu anggota mencapai milestone. Dari First Blood untuk TYFCB pertama, sampai Centurion saat skor pribadi melewati seratus.",
	},
	{
		ID: "admin", Route: "/admin/tyfcb",
		Title: "Verifikasi admin",
		Body:  "Growth Coordinator memverifikasi setiap TYFCB. Saat disetujui, sistem menulis poin ke ledger secara permanen — koreksi ditulis sebagai baris baru bertanda negatif, bukan menghapus.",
		Voice: "Ini sisi admin. Growth Coordinator memverifikasi setiap TYFCB. Saat disetujui, sistem menulis poin ke ledger secara permanen. Koreksi ditulis sebagai baris baru bertanda negatif, bukan menghapus.",
	},
	{
		ID: "done", Route: "/",
		Title: "Tur selesai",
		Body:  "Itu alur lengkap BRAG 2026. Selamat mencoba.",
		Voice: "Itu alur lengkap BRAG 2026. Selamat mencoba.",
	},
}

func (s *Server) handleTourSteps(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, tourSteps)
}

// Narration is identical on every playback, so each clip is synthesised once
// per process lifetime and replayed from memory afterwards.
var (
	tourAudioMu    sync.RWMutex
	tourAudioCache = map[string][]byte{}
)

const elevenLabsModel = "eleven_multilingual_v2"

// handleTourVoice proxies ElevenLabs so the API key stays server-side. Any
// non-audio reply tells the client to fall back to browser speech synthesis,
// which keeps the tour from ever being silent.
func (s *Server) handleTourVoice(w http.ResponseWriter, r *http.Request) {
	stepID := r.URL.Query().Get("step")

	var step *TourStep
	for i := range tourSteps {
		if tourSteps[i].ID == stepID {
			step = &tourSteps[i]
			break
		}
	}
	if step == nil {
		writeError(w, http.StatusNotFound, "Langkah tur tidak dikenal.")
		return
	}

	if s.cfg.ElevenLabsKey == "" || s.cfg.ElevenLabsVoice == "" {
		// No credentials configured: the tour narrates in the browser.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	tourAudioMu.RLock()
	cached, ok := tourAudioCache[step.ID]
	tourAudioMu.RUnlock()

	if ok {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("X-Cache", "hit")
		_, _ = w.Write(cached)
		return
	}

	payload, _ := json.Marshal(map[string]any{
		"text":     step.Voice,
		"model_id": elevenLabsModel,
		"voice_settings": map[string]float64{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	})

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost,
		"https://api.elevenlabs.io/v1/text-to-speech/"+s.cfg.ElevenLabsVoice,
		bytes.NewReader(payload))
	if err != nil {
		fail(w, err)
		return
	}
	req.Header.Set("xi-api-key", s.cfg.ElevenLabsKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("elevenlabs request failed", "err", err)
		writeError(w, http.StatusBadGateway, "Text-to-speech tidak tersedia.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		slog.Error("elevenlabs rejected the request", "status", resp.StatusCode, "detail", string(detail))

		if resp.StatusCode == http.StatusPaymentRequired {
			slog.Error("hint: free ElevenLabs plans cannot use library or professional " +
				"voices via the API. Set ELEVENLABS_VOICE_ID to a premade voice, or " +
				"upgrade the plan. The tour uses the browser voice meanwhile.")
		}

		writeError(w, http.StatusBadGateway, "Text-to-speech tidak tersedia.")
		return
	}

	// Capped so a malformed upstream reply cannot exhaust memory.
	audio, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		fail(w, err)
		return
	}

	tourAudioMu.Lock()
	tourAudioCache[step.ID] = audio
	tourAudioMu.Unlock()

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("X-Cache", "miss")
	_, _ = w.Write(audio)
}
