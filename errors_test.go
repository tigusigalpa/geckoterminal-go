package geckoterminal_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func TestAPIErrorError(t *testing.T) {
	t.Parallel()
	e := &geckoterminal.APIError{
		StatusCode: 404,
		Method:     "GET",
		Path:       "/api/v2/networks/x",
		Errors:     []geckoterminal.ErrorDetail{{Status: "404", Title: "Not Found"}},
		RefID:      "ref-1",
		RequestID:  "req-1",
	}
	got := e.Error()
	for _, want := range []string{"GET", "/api/v2/networks/x", "404", "Not Found", "ref-1", "req-1"} {
		if !contains(got, want) {
			t.Fatalf("Error() = %q, missing %q", got, want)
		}
	}
}

func TestAPIErrorErrorNoDetails(t *testing.T) {
	t.Parallel()
	e := &geckoterminal.APIError{StatusCode: 503, Method: "GET", Path: "/x"}
	if got := e.Error(); !contains(got, "503") {
		t.Fatalf("Error() = %q", got)
	}
}

func TestAPIErrorUnwrap(t *testing.T) {
	t.Parallel()
	e := &geckoterminal.APIError{StatusCode: http.StatusNotFound}
	if !errors.Is(e, geckoterminal.ErrNotFound) {
		t.Fatal("404 does not unwrap to ErrNotFound")
	}
	e = &geckoterminal.APIError{StatusCode: http.StatusTooManyRequests}
	if !errors.Is(e, geckoterminal.ErrRateLimited) {
		t.Fatal("429 does not unwrap to ErrRateLimited")
	}
	e = &geckoterminal.APIError{StatusCode: http.StatusInternalServerError}
	if errors.Is(e, geckoterminal.ErrNotFound) || errors.Is(e, geckoterminal.ErrRateLimited) {
		t.Fatal("500 unwraps to a sentinel")
	}
}

func TestIsNotFoundIsRateLimited(t *testing.T) {
	t.Parallel()
	if geckoterminal.IsNotFound(nil) || geckoterminal.IsRateLimited(nil) {
		t.Fatal("nil error")
	}
	if geckoterminal.IsNotFound(errors.New("x")) {
		t.Fatal("plain error is not found")
	}
}

func TestValidationError(t *testing.T) {
	t.Parallel()
	e := &geckoterminal.ValidationError{Field: "page", Value: "11", Message: "must be between 1 and 10"}
	got := e.Error()
	for _, want := range []string{"page", `"11"`, "between 1 and 10"} {
		if !contains(got, want) {
			t.Fatalf("Error() = %q, missing %q", got, want)
		}
	}
	if !errors.Is(e, geckoterminal.ErrInvalidArgument) {
		t.Fatal("ValidationError does not unwrap to ErrInvalidArgument")
	}
}

func TestAPIErrorFields(t *testing.T) {
	t.Parallel()
	e := &geckoterminal.APIError{
		StatusCode: 429,
		RetryAfter: 5 * time.Second,
		Body:       []byte("x"),
	}
	if e.RetryAfter != 5*time.Second || string(e.Body) != "x" {
		t.Fatalf("fields = %+v", e)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
