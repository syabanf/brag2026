package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

func requestWith(query string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/api/admin/members?"+query, nil)
}

func TestPageFromClampsHostileInput(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"nothing supplied", "", domain.DefaultPageSize, 0},
		{"explicit limit", "limit=10", 10, 0},
		{"per_page is accepted too", "per_page=15", 15, 0},

		// The cap is the point: without it one request can pull every row.
		{"oversized limit is capped", "limit=100000", domain.MaxPageSize, 0},
		{"negative limit falls back", "limit=-5", domain.DefaultPageSize, 0},
		{"zero limit falls back", "limit=0", domain.DefaultPageSize, 0},
		{"garbage limit falls back", "limit=abc", domain.DefaultPageSize, 0},

		{"explicit offset", "limit=10&offset=30", 10, 30},
		{"negative offset is clamped", "offset=-99", domain.DefaultPageSize, 0},
		{"page is translated", "page=3&limit=10", 10, 20},
		{"page 1 starts at zero", "page=1&limit=10", 10, 0},
		{"page without limit uses the default", "page=2", domain.DefaultPageSize, domain.DefaultPageSize},
		// An explicit offset wins, so a client can page precisely.
		{"offset beats page", "page=5&offset=7&limit=10", 10, 7},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := pageFrom(requestWith(c.query))

			if got.Limit != c.wantLimit {
				t.Errorf("limit = %d, want %d", got.Limit, c.wantLimit)
			}
			if got.Offset != c.wantOffset {
				t.Errorf("offset = %d, want %d", got.Offset, c.wantOffset)
			}
		})
	}
}

// A filter must distinguish "not supplied" from "supplied as false", or
// ?converted=false would silently mean "any".
func TestBoolParam(t *testing.T) {
	cases := []struct {
		query string
		want  *bool
	}{
		{"", nil},
		{"converted=true", ptrBool(true)},
		{"converted=false", ptrBool(false)},
		{"converted=1", ptrBool(true)},
		{"converted=0", ptrBool(false)},
		// Nonsense is treated as absent rather than as false.
		{"converted=maybe", nil},
	}

	for _, c := range cases {
		got := boolParam(requestWith(c.query), "converted")

		switch {
		case c.want == nil && got != nil:
			t.Errorf("%q gave %v, want nil", c.query, *got)
		case c.want != nil && got == nil:
			t.Errorf("%q gave nil, want %v", c.query, *c.want)
		case c.want != nil && got != nil && *got != *c.want:
			t.Errorf("%q gave %v, want %v", c.query, *got, *c.want)
		}
	}
}

func TestDateParam(t *testing.T) {
	if got := dateParam(requestWith(""), "from"); got != nil {
		t.Errorf("absent date gave %v, want nil", got)
	}

	got := dateParam(requestWith("from=2026-09-15"), "from")
	if got == nil || got.Format(dateLayout) != "2026-09-15" {
		t.Errorf("parsed %v, want 2026-09-15", got)
	}

	// A filter is a convenience; failing the whole request over a stray
	// character would help nobody.
	for _, bad := range []string{"from=15-09-2026", "from=tomorrow", "from=2026-13-45"} {
		if got := dateParam(requestWith(bad), "from"); got != nil {
			t.Errorf("%q was accepted as %v", bad, got)
		}
	}
}

func TestSearchParamTrims(t *testing.T) {
	if got := searchParam(requestWith("q=%20%20budi%20%20")); got != "budi" {
		t.Errorf("got %q, want %q", got, "budi")
	}
	if got := searchParam(requestWith("q=%20%20")); got != "" {
		t.Errorf("whitespace-only search should be empty, got %q", got)
	}
}

func ptrBool(v bool) *bool { return &v }
