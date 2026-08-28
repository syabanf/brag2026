package http

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// The middleware chain is shared by every request in flight. These tests run
// it concurrently so a data race or a leaky guard shows up here rather than
// under real load. Run with -race.

// The limiter is one shared counter. Under a burst it must still admit
// exactly the configured number — over-admitting is the failure that matters,
// since that is what a password-guessing client is counting on.
func TestRateLimiterUnderBurst(t *testing.T) {
	const limit = 20
	const attempts = 200

	r := chi.NewRouter()
	r.With(httprate.LimitByIP(limit, time.Minute)).Post("/login", ok.ServeHTTP)

	var admitted, blocked atomic.Int64
	var start sync.WaitGroup
	var done sync.WaitGroup

	start.Add(1)
	done.Add(attempts)

	for i := 0; i < attempts; i++ {
		go func() {
			defer done.Done()
			start.Wait()

			req := httptest.NewRequest(http.MethodPost, "/login", nil)
			req.RemoteAddr = "203.0.113.99:5000"
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			switch rec.Code {
			case http.StatusOK:
				admitted.Add(1)
			case http.StatusTooManyRequests:
				blocked.Add(1)
			default:
				t.Errorf("unexpected status %d", rec.Code)
			}
		}()
	}

	start.Done()
	done.Wait()

	if got := admitted.Load(); got > limit {
		t.Errorf("admitted %d of %d requests, the limit is %d", got, attempts, limit)
	}
	if admitted.Load()+blocked.Load() != attempts {
		t.Errorf("%d requests went unaccounted for",
			attempts-admitted.Load()-blocked.Load())
	}
	if blocked.Load() == 0 {
		t.Error("nothing was blocked; the limiter did not engage")
	}
}

// Many clients at once must each get their own budget, and one client's burst
// must not spill into another's.
func TestRateLimiterKeepsClientsSeparateUnderLoad(t *testing.T) {
	const clients = 50
	const limit = 3

	r := chi.NewRouter()
	r.With(httprate.LimitByIP(limit, time.Minute)).Post("/login", ok.ServeHTTP)

	admitted := make([]atomic.Int64, clients)
	var wg sync.WaitGroup

	for c := 0; c < clients; c++ {
		for i := 0; i < limit*2; i++ {
			wg.Add(1)
			go func(c int) {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodPost, "/login", nil)
				// Distinct /24 hosts, so no two clients share a bucket.
				req.RemoteAddr = "198.51.100." + itoa(c) + ":5000"
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				if rec.Code == http.StatusOK {
					admitted[c].Add(1)
				}
			}(c)
		}
	}
	wg.Wait()

	for c := 0; c < clients; c++ {
		got := admitted[c].Load()
		if got == 0 {
			t.Errorf("client %d was shut out entirely", c)
		}
		if got > limit {
			t.Errorf("client %d got %d of a %d budget", c, got, limit)
		}
	}
}

// Role guards read from the request context, which is per-request. This runs
// every role at once to prove no request can pick up another's identity.
func TestRoleGuardsUnderConcurrentMixedRoles(t *testing.T) {
	const rounds = 200

	cases := []struct {
		role domain.Role
		want int
	}{
		{domain.RoleMember, http.StatusForbidden},
		{domain.RoleCaptain, http.StatusForbidden},
		{domain.RoleAdmin, http.StatusOK},
		{"", http.StatusUnauthorized},
	}

	var wg sync.WaitGroup
	for i := 0; i < rounds; i++ {
		for _, c := range cases {
			wg.Add(1)
			go func(role domain.Role, want int) {
				defer wg.Done()
				rec := httptest.NewRecorder()
				requireAdmin(ok).ServeHTTP(rec, requestAs(role))
				if rec.Code != want {
					t.Errorf("role %q got %d, want %d", role, rec.Code, want)
				}
			}(c.role, c.want)
		}
	}
	wg.Wait()
}

// secureHeaders writes to a header map per response. Concurrent traffic must
// not drop any of them.
func TestSecureHeadersUnderLoad(t *testing.T) {
	const rounds = 500

	var wg sync.WaitGroup
	for i := 0; i < rounds; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			secureHeaders(ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))
			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Error("a response went out without its headers")
			}
		}()
	}
	wg.Wait()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
