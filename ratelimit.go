package geckoterminal

import (
	"context"
	"sync"
	"time"
)

// RateLimit configures the client-side token-bucket rate limiter.
type RateLimit struct {
	// RequestsPerMinute is the sustained request rate. A value <= 0
	// disables the limiter.
	RequestsPerMinute float64
	// Burst is the maximum number of requests that may be sent
	// instantly before the sustained rate applies. A value <= 0 is
	// treated as 1.
	Burst int
}

var (
	// DefaultRateLimit matches the public GeckoTerminal API allowance
	// of roughly 10 requests per minute.
	DefaultRateLimit = RateLimit{RequestsPerMinute: 10, Burst: 10}
	// DefaultProRateLimit matches the CoinGecko Pro on-chain allowance
	// of roughly 250 requests per minute.
	DefaultProRateLimit = RateLimit{RequestsPerMinute: 250, Burst: 50}
)

// rateLimiter is a lazily refilled token bucket. It launches no
// background goroutines and is safe for concurrent use.
type rateLimiter struct {
	mu     sync.Mutex
	rate   float64 // tokens per second; <= 0 disables the limiter
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
	sleep  func(ctx context.Context, d time.Duration) error
}

// newRateLimiter builds a rateLimiter for the given RateLimit. now may
// be nil, in which case time.Now is used.
func newRateLimiter(rl RateLimit, now func() time.Time) *rateLimiter {
	if now == nil {
		now = time.Now
	}
	burst := rl.Burst
	if burst <= 0 {
		burst = 1
	}
	return &rateLimiter{
		rate:   rl.RequestsPerMinute / 60,
		burst:  float64(burst),
		tokens: float64(burst),
		last:   now(),
		now:    now,
		sleep:  sleepContext,
	}
}

// Wait blocks until a token is available or ctx is done. When the
// limiter is disabled it only checks the context and returns
// immediately.
func (l *rateLimiter) Wait(ctx context.Context) error {
	if l == nil || l.rate <= 0 {
		return ctx.Err()
	}
	for {
		l.mu.Lock()
		now := l.now()
		l.tokens += now.Sub(l.last).Seconds() * l.rate
		if l.tokens > l.burst {
			l.tokens = l.burst
		}
		l.last = now
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		missing := 1 - l.tokens
		wait := time.Duration(missing / l.rate * float64(time.Second))
		l.mu.Unlock()
		if err := l.sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// sleepContext sleeps for d or until ctx is done, whichever happens
// first.
func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
