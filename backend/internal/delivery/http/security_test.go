package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// ok is the handler under test in the middleware cases: it succeeds unless the
// middleware in front of it stops the request first.
var ok = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestSecureHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	secureHeaders(ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	// Each of these closes a specific hole, so a missing one is a regression
	// rather than a style lapse.
	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Cache-Control":          "no-store",
	}

	for header, value := range want {
		if got := rec.Header().Get(header); got != value {
			t.Errorf("%s = %q, want %q", header, got, value)
		}
	}

	if !strings.Contains(rec.Header().Get("Permissions-Policy"), "camera=()") {
		t.Error("Permissions-Policy should deny camera")
	}
}

func TestLimitBodyRejectsOversizedPayloads(t *testing.T) {
	// A handler that reads the body, which is where the cap actually bites.
	reader := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	cases := []struct {
		name string
		size int
		want int
	}{
		{"within the cap", 1024, http.StatusOK},
		{"at the cap", maxBodyBytes, http.StatusOK},
		{"over the cap", maxBodyBytes + 1, http.StatusRequestEntityTooLarge},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := bytes.Repeat([]byte("a"), c.size)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			limitBody(reader).ServeHTTP(rec, req)

			if rec.Code != c.want {
				t.Errorf("%d bytes gave %d, want %d", c.size, rec.Code, c.want)
			}
		})
	}
}

// The login limiter is what stands between an exposed login form and a
// password list.
func TestLoginRateLimit(t *testing.T) {
	r := chi.NewRouter()
	r.With(httprate.LimitByIP(10, time.Minute)).Post("/login", ok.ServeHTTP)

	var lastCode int
	for i := 1; i <= 12; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "203.0.113.10:5000"
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		lastCode = rec.Code

		if i <= 10 && rec.Code != http.StatusOK {
			t.Fatalf("attempt %d was blocked at %d; the limit should allow ten", i, rec.Code)
		}
		if i == 11 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("attempt 11 gave %d, want 429", rec.Code)
		}
	}

	if lastCode != http.StatusTooManyRequests {
		t.Errorf("still open after 12 attempts (%d)", lastCode)
	}
}

// One client hitting the limit must not lock out everyone else.
func TestRateLimitIsPerClient(t *testing.T) {
	r := chi.NewRouter()
	r.With(httprate.LimitByIP(3, time.Minute)).Post("/login", ok.ServeHTTP)

	exhaust := func(ip string) int {
		var code int
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodPost, "/login", nil)
			req.RemoteAddr = ip + ":5000"
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			code = rec.Code
		}
		return code
	}

	if got := exhaust("203.0.113.20"); got != http.StatusTooManyRequests {
		t.Fatalf("first client was not limited: %d", got)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "203.0.113.21:5000"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("second client got %d; the limit must be per client", rec.Code)
	}
}

// ── Role guards ───────────────────────────────────────────────────────────

func requestAs(role domain.Role) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/members", nil)
	if role == "" {
		return req
	}
	user := &domain.User{ID: "user-1", Role: role}
	return req.WithContext(context.WithValue(req.Context(), ctxUser, user))
}

func TestRoleGuards(t *testing.T) {
	cases := []struct {
		name  string
		guard func(http.Handler) http.Handler
		role  domain.Role
		want  int
	}{
		{"anonymous is turned away", requireAuth, "", http.StatusUnauthorized},
		{"any member may pass requireAuth", requireAuth, domain.RoleMember, http.StatusOK},

		{"anonymous cannot reach admin", requireAdmin, "", http.StatusUnauthorized},
		{"a member cannot reach admin", requireAdmin, domain.RoleMember, http.StatusForbidden},
		// A captain is not an admin: they run their own team, not the season.
		{"a captain cannot reach admin", requireAdmin, domain.RoleCaptain, http.StatusForbidden},
		{"an admin can", requireAdmin, domain.RoleAdmin, http.StatusOK},

		{"a member cannot reach captain routes", requireCaptain, domain.RoleMember, http.StatusForbidden},
		{"a captain can", requireCaptain, domain.RoleCaptain, http.StatusOK},
		// An admin can do anything a captain can.
		{"an admin can too", requireCaptain, domain.RoleAdmin, http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c.guard(ok).ServeHTTP(rec, requestAs(c.role))

			if rec.Code != c.want {
				t.Errorf("got %d, want %d", rec.Code, c.want)
			}
		})
	}
}

// A rejected request must not leak which resource it was asking about.
func TestRoleGuardsReturnAGenericMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	requireAdmin(ok).ServeHTTP(rec, requestAs(domain.RoleMember))

	body := rec.Body.String()
	if strings.Contains(body, "members") || strings.Contains(body, "user-1") {
		t.Errorf("the refusal echoes request detail: %s", body)
	}
}

// ── Error mapping ─────────────────────────────────────────────────────────

func TestFailMapsDomainErrorsToStatusCodes(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{domain.Invalid("nilai tidak valid"), http.StatusBadRequest},
		{domain.NotFound("tidak ada"), http.StatusNotFound},
		{domain.Forbidden("bukan tim Anda"), http.StatusForbidden},
		{domain.Conflict("sudah berubah"), http.StatusConflict},
		{domain.NewError(domain.ErrUnauthorized, "masuk dulu"), http.StatusUnauthorized},
		{domain.NewError(domain.ErrAlreadyExists, "sudah ada"), http.StatusConflict},
		// Bare sentinels map too, for callers that return them directly.
		{domain.ErrNotFound, http.StatusNotFound},
		{domain.ErrForbidden, http.StatusForbidden},
	}

	for _, c := range cases {
		rec := httptest.NewRecorder()
		fail(rec, c.err)

		if rec.Code != c.want {
			t.Errorf("%v gave %d, want %d", c.err, rec.Code, c.want)
		}
	}
}

// An unexpected failure must become a generic 500 rather than handing the
// caller a database error to read.
func TestFailHidesUnexpectedErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	fail(rec, errors.New(`pq: relation "app_users" does not exist`))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "app_users") {
		t.Errorf("internal detail leaked to the client: %s", rec.Body.String())
	}
}

// A domain error's message is written for the user, so it should survive.
func TestFailKeepsUserFacingMessages(t *testing.T) {
	rec := httptest.NewRecorder()
	fail(rec, domain.Invalid("Nilai transaksi tidak valid."))

	if !strings.Contains(rec.Body.String(), "Nilai transaksi tidak valid.") {
		t.Errorf("the message was dropped: %s", rec.Body.String())
	}
}
