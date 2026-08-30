package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/ikurniawann/brag2026/backend/internal/config"
	"github.com/ikurniawann/brag2026/backend/internal/delivery/http/apidocs"
)

// Documentation drifts the moment a route is added and nobody remembers the
// spec. These tests make that a failing build instead of a support question.

func routerUnderTest(t *testing.T) chi.Router {
	t.Helper()

	s := NewServer(Deps{Config: &config.Config{
		CookieName:     "brag_session",
		AllowedOrigins: []string{"http://localhost:5173"},
	}})

	router, ok := s.Router().(chi.Router)
	if !ok {
		t.Fatal("the server no longer exposes a chi router")
	}
	return router
}

// walkRoutes lists every method and path the server actually serves.
func walkRoutes(t *testing.T) map[string]bool {
	t.Helper()

	routes := map[string]bool{}
	err := chi.Walk(routerUnderTest(t),
		func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
			// chi reports a mounted sub-router with a trailing wildcard; the
			// docs are one page, not an endpoint per file it can serve.
			if strings.HasPrefix(route, "/api/docs") {
				return nil
			}
			routes[strings.ToUpper(method)+" "+strings.TrimSuffix(route, "/")] = true
			return nil
		})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return routes
}

func documentedOperations(t *testing.T) map[string]bool {
	t.Helper()

	spec := apidocs.Load()
	if len(spec.Paths) == 0 {
		t.Fatal("the specification has no paths")
	}

	out := map[string]bool{}
	for path, methods := range spec.Paths {
		for method := range methods {
			out[strings.ToUpper(method)+" "+path] = true
		}
	}
	return out
}

// Every route the server answers must appear in the specification. A missing
// one is an endpoint nobody outside the team can discover.
func TestEveryRouteIsDocumented(t *testing.T) {
	documented := documentedOperations(t)

	var missing []string
	for route := range walkRoutes(t) {
		if !documented[route] {
			missing = append(missing, route)
		}
	}

	if len(missing) > 0 {
		t.Errorf("%d route(s) are served but not in openapi.yaml:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}

// And nothing is documented that the server does not serve — a promise the
// API cannot keep is worse than no promise.
func TestNothingIsDocumentedThatDoesNotExist(t *testing.T) {
	served := walkRoutes(t)

	var phantom []string
	for route := range documentedOperations(t) {
		if !served[route] {
			phantom = append(phantom, route)
		}
	}

	if len(phantom) > 0 {
		t.Errorf("%d documented operation(s) have no route:\n  %s",
			len(phantom), strings.Join(phantom, "\n  "))
	}
}

// The page is rendered from the specification, so a template that fails to
// execute would ship as a blank reference.
func TestDocsPageRenders(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	apidocs.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	body := rec.Body.String()
	for _, want := range []string{"<html", "Referensi API", "openapi.yaml", "openapi.json"} {
		if !strings.Contains(body, want) {
			t.Errorf("the page is missing %q", want)
		}
	}
	// An unrendered action means the template silently dropped content.
	if strings.Contains(body, "{{") {
		t.Error("the page contains an unrendered template action")
	}
}

func TestSpecIsServedInBothFormats(t *testing.T) {
	for _, c := range []struct{ path, contentType string }{
		{"/openapi.yaml", "application/yaml"},
		{"/openapi.json", "application/json"},
	} {
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, c.path, nil)
		apidocs.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("%s gave %d", c.path, rec.Code)
		}
		if !strings.HasPrefix(rec.Header().Get("Content-Type"), c.contentType) {
			t.Errorf("%s served %q", c.path, rec.Header().Get("Content-Type"))
		}
		if rec.Body.Len() == 0 {
			t.Errorf("%s served nothing", c.path)
		}
	}
}
