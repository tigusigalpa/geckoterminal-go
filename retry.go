package geckoterminal

import (
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// RetryPolicy controls how failed requests are retried.
//
// All endpoints of the GeckoTerminal API are idempotent GET requests, so
// retrying on transient failures is safe. The zero value is not usable
// directly; pass the policy to WithRetryPolicy and unset fields are
// normalized to DefaultRetryPolicy values.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retries after the first
	// attempt. Zero disables retries.
	MaxRetries int
	// BaseDelay is the delay before the first retry. The delay doubles
	// on every subsequent attempt.
	BaseDelay time.Duration
	// MaxDelay caps the exponential backoff delay.
	MaxDelay time.Duration
	// Jitter is the fraction (0..1) of the delay randomized up or down.
	Jitter float64
	// RetryOn lists HTTP status codes that trigger a retry.
	RetryOn []int
}

// DefaultRetryPolicy is used when WithRetryPolicy is not provided or
// when individual fields of a custom policy are left unset.
var DefaultRetryPolicy = RetryPolicy{
	MaxRetries: 3,
	BaseDelay:  time.Second,
	MaxDelay:   30 * time.Second,
	Jitter:     0.2,
	RetryOn:    []int{429, 500, 502, 503, 504},
}

// normalize fills unset fields with the DefaultRetryPolicy values and
// clamps Jitter into [0, 1].
func (p RetryPolicy) normalize() RetryPolicy {
	if p.BaseDelay <= 0 {
		p.BaseDelay = DefaultRetryPolicy.BaseDelay
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = DefaultRetryPolicy.MaxDelay
	}
	if p.Jitter < 0 {
		p.Jitter = 0
	}
	if p.Jitter > 1 {
		p.Jitter = 1
	}
	if len(p.RetryOn) == 0 {
		p.RetryOn = append([]int(nil), DefaultRetryPolicy.RetryOn...)
	}
	return p
}

// retryable reports whether the given HTTP status code should be retried.
func (p RetryPolicy) retryable(status int) bool {
	for _, s := range p.RetryOn {
		if s == status {
			return true
		}
	}
	return false
}

// delay computes the backoff for the given retry attempt (1-based). A
// server-provided Retry-After value takes precedence over the computed
// exponential backoff.
func (p RetryPolicy) delay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}
	d := p.BaseDelay
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= p.MaxDelay {
			d = p.MaxDelay
			break
		}
	}
	if d > p.MaxDelay {
		d = p.MaxDelay
	}
	if p.Jitter > 0 {
		// shift delay by +/- Jitter*delay, uniform random.
		delta := p.Jitter * float64(d)
		d = time.Duration(float64(d) - delta + 2*delta*rand.Float64()) //nolint:gosec // jitter does not need crypto-strength randomness
	}
	return d
}

// parseRetryAfter parses the Retry-After header value, which is either a
// non-negative number of seconds or an HTTP-date. It returns 0 when the
// value is missing, malformed or already in the past.
func parseRetryAfter(header string, now time.Time) time.Duration {
	if header == "" {
		return 0
	}
	if secs, err := strconv.Atoi(header); err == nil {
		if secs < 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(header); err == nil {
		if d := t.Sub(now); d > 0 {
			return d
		}
	}
	return 0
}
