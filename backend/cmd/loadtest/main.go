// Command loadtest drives the API with concurrent traffic and reports latency
// percentiles per endpoint.
//
// It is a measuring tool, not a test: it needs a running server with seeded
// data, so it lives outside `go test` and is invoked from scripts/stress.sh.
//
//	go run ./cmd/loadtest -base http://localhost:8081 -workers 50 -duration 30s
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"
)

type target struct {
	name string
	path string
	// weight approximates how often the real UI calls this endpoint, so the
	// mix resembles a busy afternoon rather than a uniform sweep.
	weight int
}

// The read paths a logged-in admin actually loads. The leaderboard and the
// dashboard are the expensive ones — both aggregate the whole ledger.
var targets = []target{
	{"health", "/api/health", 1},
	{"leaderboard", "/api/leaderboard", 6},
	{"dashboard", "/api/dashboard", 6},
	{"activity", "/api/activity?limit=20", 3},
	{"badges", "/api/badges", 2},
	{"prizes", "/api/prizes", 2},
	{"raffle tickets", "/api/raffle/tickets", 2},
	{"events", "/api/events", 1},
	{"tyfcb (paged)", "/api/admin/tyfcb?limit=25", 5},
	{"tyfcb (filtered)", "/api/admin/tyfcb?limit=25&status=verified&q=a", 3},
	{"visitors (paged)", "/api/admin/visitors?limit=25", 4},
	{"visitors (filtered)", "/api/admin/visitors?limit=25&converted=true&q=a", 3},
	{"members (paged)", "/api/admin/members?limit=25", 4},
	{"members (search)", "/api/admin/members?limit=25&q=a", 3},
	{"deep page", "/api/admin/tyfcb?limit=25&offset=150", 2},
}

type sample struct {
	target  int
	latency time.Duration
	status  int
	err     bool
}

func main() {
	base := flag.String("base", "http://localhost:8081", "API base URL")
	email := flag.String("email", "demo.admin@brag2026.id", "account to sign in as")
	password := flag.String("password", "demo123", "password")
	workers := flag.Int("workers", 50, "concurrent clients")
	duration := flag.Duration("duration", 30*time.Second, "how long to run")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	budget := flag.Duration("budget", 0, "fail if p95 of any endpoint exceeds this")
	flag.Parse()

	client, err := signIn(*base, *email, *password, *timeout)
	if err != nil {
		log.Fatalf("sign in: %v", err)
	}

	fmt.Printf("── load: %d workers for %s against %s ──\n\n", *workers, *duration, *base)

	// One slot per (worker, target) pair to build the rotation; each worker
	// walks its own offset so the mix stays even without coordination.
	rotation := make([]int, 0, 64)
	for i, t := range targets {
		for w := 0; w < t.weight; w++ {
			rotation = append(rotation, i)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	samples := make(chan sample, 4096)
	var wg sync.WaitGroup

	started := time.Now()
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for n := 0; ctx.Err() == nil; n++ {
				idx := rotation[(offset+n)%len(rotation)]
				s := fetch(ctx, client, *base+targets[idx].path)
				s.target = idx

				// A request cut short by the run deadline says nothing about
				// the server, so it is dropped rather than counted as a
				// failure the report would blame on the API.
				if s.err && ctx.Err() != nil {
					return
				}

				select {
				case samples <- s:
				case <-ctx.Done():
					return
				}
			}
		}(w)
	}

	go func() { wg.Wait(); close(samples) }()

	collected := make([][]time.Duration, len(targets))
	failures := make([]int, len(targets))
	statuses := map[int]int{}
	total := 0

	for s := range samples {
		total++
		statuses[s.status]++
		if s.err || s.status >= 400 {
			failures[s.target]++
			continue
		}
		collected[s.target] = append(collected[s.target], s.latency)
	}
	elapsed := time.Since(started)

	os.Exit(report(collected, failures, statuses, total, elapsed, *budget))
}

// signIn logs in once and keeps the session cookie. Every worker shares it:
// the login route is rate limited, and a real crowd is already signed in.
func signIn(base, email, password string, timeout time.Duration) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: timeout,
		Transport: &http.Transport{
			// Without a pool this large, workers spend the run in TCP setup
			// and the numbers measure the dialer rather than the server.
			MaxIdleConns:        512,
			MaxIdleConnsPerHost: 512,
			MaxConnsPerHost:     512,
		},
	}

	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := client.Post(base+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d — is the demo data seeded?", resp.StatusCode)
	}

	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if len(jar.Cookies(u)) == 0 {
		return nil, fmt.Errorf("no session cookie was set")
	}
	return client, nil
}

func fetch(ctx context.Context, client *http.Client, url string) sample {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return sample{err: true}
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return sample{latency: time.Since(start), err: true}
	}
	defer resp.Body.Close()

	// Drain the body: latency without it measures headers, not the response.
	_, _ = io.Copy(io.Discard, resp.Body)
	return sample{latency: time.Since(start), status: resp.StatusCode}
}

func report(
	collected [][]time.Duration, failures []int, statuses map[int]int,
	total int, elapsed time.Duration, budget time.Duration,
) int {
	fmt.Printf("%-22s %8s %7s %9s %9s %9s %9s\n",
		"endpoint", "reqs", "fail", "p50", "p95", "p99", "max")
	fmt.Println(rule(80))

	exit := 0
	for i, t := range targets {
		lat := collected[i]
		if len(lat) == 0 && failures[i] == 0 {
			continue
		}
		sort.Slice(lat, func(a, b int) bool { return lat[a] < lat[b] })

		fmt.Printf("%-22s %8d %7d %9s %9s %9s %9s\n",
			t.name, len(lat)+failures[i], failures[i],
			ms(pct(lat, 50)), ms(pct(lat, 95)), ms(pct(lat, 99)), ms(pct(lat, 100)))

		if budget > 0 && pct(lat, 95) > budget {
			exit = 1
		}
		if failures[i] > 0 {
			exit = 1
		}
	}

	fmt.Println(rule(80))
	fmt.Printf("\n%d requests in %s — %.0f req/s\n",
		total, elapsed.Round(time.Millisecond), float64(total)/elapsed.Seconds())

	fmt.Print("statuses: ")
	codes := make([]int, 0, len(statuses))
	for code := range statuses {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		label := fmt.Sprint(code)
		if code == 0 {
			label = "transport error"
		}
		fmt.Printf("%s×%d  ", label, statuses[code])
	}
	fmt.Println()

	if exit != 0 {
		fmt.Println("\nFAIL — see the failures column, or a p95 over budget")
	} else if budget > 0 {
		fmt.Printf("\nPASS — every endpoint stayed under a %s p95\n", budget)
	}
	return exit
}

// pct returns the p-th percentile of a sorted slice, using nearest-rank so a
// small sample reports a value that was actually observed.
func pct(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	i := (p * len(sorted)) / 100
	if i >= len(sorted) {
		i = len(sorted) - 1
	}
	return sorted[i]
}

func ms(d time.Duration) string {
	if d == 0 {
		return "–"
	}
	return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
}

func rule(n int) string {
	b := make([]byte, 0, n*3)
	for i := 0; i < n; i++ {
		b = append(b, "─"...)
	}
	return string(b)
}
