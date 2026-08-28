package domain

import "errors"

// Transport-agnostic failures. The HTTP layer maps these onto status codes so
// use cases never import net/http.
var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("authentication required")
	ErrForbidden     = errors.New("insufficient permissions")
	ErrConflict      = errors.New("conflicting state")
	ErrAlreadyExists = errors.New("resource already exists")
)

// Error carries a human-readable message (shown to the user, in Indonesian)
// alongside the sentinel that decides the status code.
type Error struct {
	Kind    error
	Message string
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Kind }

func NewError(kind error, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

func Invalid(message string) *Error   { return NewError(ErrInvalidInput, message) }
func NotFound(message string) *Error  { return NewError(ErrNotFound, message) }
func Forbidden(message string) *Error { return NewError(ErrForbidden, message) }
func Conflict(message string) *Error  { return NewError(ErrConflict, message) }
