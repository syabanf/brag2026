// Package http adapts the use cases to JSON over HTTP. It is the only layer
// that knows about status codes, cookies and request shapes.
package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ikurniawann/brag2026/backend/internal/domain"
)

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}

// fail turns a domain error into the right status code. Anything unrecognised
// is a 500 and is logged rather than leaked to the caller.
func fail(w http.ResponseWriter, err error) {
	var domErr *domain.Error
	if errors.As(err, &domErr) {
		writeError(w, statusFor(domErr.Kind), domErr.Message)
		return
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "Data tidak ditemukan.")
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "Permintaan tidak valid.")
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "Silakan masuk terlebih dahulu.")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "Akses ditolak.")
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "Terjadi konflik data.")
	default:
		slog.Error("unhandled error", "err", err)
		writeError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server.")
	}
}

func statusFor(kind error) int {
	switch {
	case errors.Is(kind, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(kind, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(kind, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(kind, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(kind, domain.ErrConflict), errors.Is(kind, domain.ErrAlreadyExists):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// decode reads a JSON body and rejects unknown fields, so a typo in a client
// payload surfaces as a 400 instead of being silently ignored.
func decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return domain.Invalid("Format permintaan tidak valid.")
	}
	return nil
}
