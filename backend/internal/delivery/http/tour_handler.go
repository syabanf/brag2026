package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
// and replayed afterwards — from memory within a process, and from disk across
// restarts.
//
// The disk half is not an optimisation. A full run of the tour is about 1,450
// characters and a free ElevenLabs plan allows 10,000 a month, so an
// in-memory-only cache buys six playthroughs and then every redeploy spends
// the quota again. Written down, the nine clips cost 1,450 characters once and
// never again.
var (
	tourAudioMu    sync.RWMutex
	tourAudioCache = map[string][]byte{}
)

// tourAudioDir is where clips are kept between restarts. Empty disables the
// disk half and leaves the in-memory cache, which is what a test wants.
var tourAudioDir = os.Getenv("TOUR_AUDIO_DIR")

// clipPath names a step's file. The id comes from tourSteps in this package,
// never from the request, but it is checked anyway: a cache key that reaches
// the filesystem is worth being certain about.
func clipPath(stepID string) string {
	if tourAudioDir == "" || !safeStepID(stepID) {
		return ""
	}
	return filepath.Join(tourAudioDir, stepID+".mp3")
}

func safeStepID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

// loadClip returns a previously written clip, or nil when there is none.
func loadClip(stepID string) []byte {
	path := clipPath(stepID)
	if path == "" {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return body
}

// storeClip writes a clip for future restarts. A failure here costs quota on
// the next boot but nothing today, so it is logged rather than returned.
func storeClip(stepID string, audio []byte) {
	path := clipPath(stepID)
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		slog.Warn("tour audio cache unwritable", "err", err)
		return
	}
	// Written beside the target and renamed, so a crash mid-write cannot
	// leave a truncated clip that would be served as if it were whole.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, audio, 0o644); err != nil {
		slog.Warn("tour audio cache write failed", "err", err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		slog.Warn("tour audio cache rename failed", "err", err)
		_ = os.Remove(tmp)
	}
}

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

	if !ok {
		// Nothing in memory: this process may still have the clip from an
		// earlier one. Checking costs a file read; not checking costs quota.
		if fromDisk := loadClip(step.ID); fromDisk != nil {
			tourAudioMu.Lock()
			tourAudioCache[step.ID] = fromDisk
			tourAudioMu.Unlock()
			cached, ok = fromDisk, true
		}
	}

	if ok {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Cache-Control", "public, max-age=86400")
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
	storeClip(step.ID, audio)

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Cache", "miss")
	_, _ = w.Write(audio)
}
