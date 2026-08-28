package http

import (
	"net/http"
)

// maxBodyBytes caps request bodies. Every endpoint here takes a small JSON
// object, so a megabyte is generous — without a cap a single client can make
// the server buffer unbounded input.
const maxBodyBytes = 1 << 20

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// secureHeaders sets the defensive headers a JSON API should always send.
// HSTS is deliberately left to the TLS terminator, which knows whether the
// deployment is actually HTTPS.
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()

		// Stops a browser from guessing a different content type than we sent.
		h.Set("X-Content-Type-Options", "nosniff")
		// The API renders no HTML, so nothing here should ever be framed.
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		// A JSON API needs no ambient capabilities.
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// Responses are per-session; a shared cache must not reuse them.
		h.Set("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}
