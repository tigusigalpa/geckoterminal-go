package geckoterminal

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	// ErrNotFound indicates a 404 response.
	ErrNotFound = errors.New("geckoterminal: resource not found")
	// ErrRateLimited indicates a 429 response.
	ErrRateLimited = errors.New("geckoterminal: rate limited")
	// ErrInvalidArgument indicates a client-side validation failure.
	ErrInvalidArgument = errors.New("geckoterminal: invalid argument")
)

// ErrorDetail is a single JSON:API error object.
type ErrorDetail struct {
	Status string `json:"status"`
	Title  string `json:"title"`
}

// APIError is returned for non-2xx responses.
//
// APIError supports errors.Is: a 404 response unwraps to ErrNotFound and
// a 429 response unwraps to ErrRateLimited. Use errors.As to inspect the
// full error, including per-endpoint details and the API request IDs.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Errors     []ErrorDetail
	RefID      string
	RequestID  string
	RetryAfter time.Duration
	// Body holds up to 64 KiB of the raw response body for diagnostics.
	// It may contain non-JSON data (for example an HTML error page from
	// a CDN edge) and never contains request credentials.
	Body []byte
}

// Error returns a short human-readable description of the API error.
func (e *APIError) Error() string {
	title := http.StatusText(e.StatusCode)
	if len(e.Errors) > 0 && e.Errors[0].Title != "" {
		title = e.Errors[0].Title
	}
	if title == "" {
		title = "error"
	}
	return fmt.Sprintf("geckoterminal: %s %s: %d %s (ref_id=%s, request_id=%s)",
		e.Method, e.Path, e.StatusCode, title, e.RefID, e.RequestID)
}

// Unwrap maps the HTTP status code to a sentinel error so that
// errors.Is(err, ErrNotFound) and errors.Is(err, ErrRateLimited) work.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return nil
	}
}

// IsNotFound reports whether err is a 404 API error.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsRateLimited reports whether err is a 429 API error.
func IsRateLimited(err error) bool { return errors.Is(err, ErrRateLimited) }

// ValidationError is returned before a request is sent when a parameter
// fails client-side validation. It unwraps to ErrInvalidArgument.
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

// Error returns a human-readable description of the validation failure.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("geckoterminal: invalid %s=%q: %s", e.Field, e.Value, e.Message)
}

// Unwrap returns ErrInvalidArgument so that errors.Is works.
func (e *ValidationError) Unwrap() error { return ErrInvalidArgument }
