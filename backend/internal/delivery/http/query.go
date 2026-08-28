package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

// pageFrom reads ?limit= and ?offset=, or ?page= as a friendlier alternative.
// Normalise clamps whatever comes back, so bad input becomes a sane page
// rather than an error the caller has to handle.
func pageFrom(r *http.Request) domain.Page {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit, _ = strconv.Atoi(q.Get("per_page"))
	}

	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset == 0 {
		if page, _ := strconv.Atoi(q.Get("page")); page > 1 {
			effective := limit
			if effective <= 0 {
				effective = domain.DefaultPageSize
			}
			offset = (page - 1) * effective
		}
	}

	return domain.Page{Limit: limit, Offset: offset}.Normalise()
}

// boolParam distinguishes "not supplied" from "supplied as false", which a
// plain bool cannot express.
func boolParam(r *http.Request, name string) *bool {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &v
}

// dateParam reads an optional YYYY-MM-DD bound. An unparseable value is
// ignored rather than rejected: a filter is a convenience, and failing the
// whole request over a stray character helps nobody.
func dateParam(r *http.Request, name string) *time.Time {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, raw)
	if err != nil {
		return nil
	}
	return &t
}

func searchParam(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("q"))
}
