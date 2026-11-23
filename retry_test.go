package geckoterminal

import (
	"net/http"
	"testing"
	"time"
)

func TestRetryPolicyNormalize(t *testing.T) {
	t.Parallel()
	p := RetryPolicy{}.normalize()
	if p.BaseDelay != DefaultRetryPolicy.BaseDelay {
		t.Fatalf("BaseDelay = %v", p.BaseDelay)
	}
	if p.MaxDelay != DefaultRetryPolicy.MaxDelay {
		t.Fatalf("MaxDelay = %v", p.MaxDelay)
	}
	if len(p.RetryOn) != len(DefaultRetryPolicy.RetryOn) {
		t.Fatalf("RetryOn = %v", p.RetryOn)
	}
	p = RetryPolicy{Jitter: -1}.normalize()
	if p.Jitter != 0 {
		t.Fatalf("Jitter = %v", p.Jitter)
	}
	p = RetryPolicy{Jitter: 2}.normalize()
	if p.Jitter != 1 {
		t.Fatalf("Jitter = %v", p.Jitter)
	}
	custom := RetryPolicy{BaseDelay: time.Second, MaxDelay: 2 * time.Second, RetryOn: []int{418}}
	n := custom.normalize()
	if n.BaseDelay != time.Second || n.MaxDelay != 2*time.Second || len(n.RetryOn) != 1 {
		t.Fatalf("custom normalize clobbered fields: %+v", n)
	}
}

func TestRetryPolicyRetryable(t *testing.T) {
	t.Parallel()
	p := DefaultRetryPolicy.normalize()
	for _, code := range []int{429, 500, 502, 503, 504} {
		if !p.retryable(code) {
			t.Fatalf("%d not retryable", code)
		}
	}
	for _, code := range []int{200, 400, 401, 403, 404, 418} {
		if p.retryable(code) {
			t.Fatalf("%d retryable", code)
		}
	}
}

func TestRetryPolicyDelay(t *testing.T) {
	t.Parallel()
	p := RetryPolicy{
		BaseDelay: time.Second,
		MaxDelay:  10 * time.Second,
		Jitter:    0,
	}.normalize()
	if d := p.delay(1, 0); d != time.Second {
		t.Fatalf("delay(1) = %v", d)
	}
	if d := p.delay(2, 0); d != 2*time.Second {
		t.Fatalf("delay(2) = %v", d)
	}
	if d := p.delay(3, 0); d != 4*time.Second {
		t.Fatalf("delay(3) = %v", d)
	}
	if d := p.delay(10, 0); d != 10*time.Second {
		t.Fatalf("delay(10) = %v, want cap 10s", d)
	}
	// Retry-After wins over backoff.
	if d := p.delay(1, 42*time.Second); d != 42*time.Second {
		t.Fatalf("delay with retryAfter = %v", d)
	}
}

func TestRetryPolicyDelayJitter(t *testing.T) {
	t.Parallel()
	p := RetryPolicy{
		BaseDelay: time.Second,
		MaxDelay:  10 * time.Second,
		Jitter:    0.5,
	}.normalize()
	for i := 0; i < 100; i++ {
		d := p.delay(1, 0)
		if d < 500*time.Millisecond || d > 1500*time.Millisecond {
			t.Fatalf("jittered delay out of bounds: %v", d)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	if d := parseRetryAfter("", now); d != 0 {
		t.Fatalf("empty = %v", d)
	}
	if d := parseRetryAfter("5", now); d != 5*time.Second {
		t.Fatalf("seconds = %v", d)
	}
	if d := parseRetryAfter("0", now); d != 0 {
		t.Fatalf("zero = %v", d)
	}
	if d := parseRetryAfter("-3", now); d != 0 {
		t.Fatalf("negative = %v", d)
	}
	future := now.Add(30 * time.Second)
	if d := parseRetryAfter(future.Format(http.TimeFormat), now); d != 30*time.Second {
		t.Fatalf("http date = %v", d)
	}
	past := now.Add(-time.Minute)
	if d := parseRetryAfter(past.Format(http.TimeFormat), now); d != 0 {
		t.Fatalf("past date = %v", d)
	}
	if d := parseRetryAfter("garbage", now); d != 0 {
		t.Fatalf("garbage = %v", d)
	}
}
