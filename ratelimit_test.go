package geckoterminal

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeClock provides deterministic time and a sleep that advances it.
type fakeClock struct {
	now    time.Time
	sleeps []time.Duration
}

func (f *fakeClock) Now() time.Time { return f.now }

func (f *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	f.sleeps = append(f.sleeps, d)
	f.now = f.now.Add(d)
	return nil
}

func newTestLimiter(rl RateLimit, fc *fakeClock) *rateLimiter {
	l := newRateLimiter(rl, fc.Now)
	l.sleep = fc.Sleep
	return l
}

func TestRateLimiterBurst(t *testing.T) {
	t.Parallel()
	fc := &fakeClock{now: time.Unix(1000, 0)}
	l := newTestLimiter(RateLimit{RequestsPerMinute: 60, Burst: 10}, fc)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("Wait %d: %v", i, err)
		}
	}
	if len(fc.sleeps) != 0 {
		t.Fatalf("unexpected sleeps %v", fc.sleeps)
	}
	// 11th request must wait for one token: 60/min = 1/sec.
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("Wait 11: %v", err)
	}
	if len(fc.sleeps) != 1 || fc.sleeps[0] != time.Second {
		t.Fatalf("sleeps = %v, want [1s]", fc.sleeps)
	}
}

func TestRateLimiterRefill(t *testing.T) {
	t.Parallel()
	fc := &fakeClock{now: time.Unix(1000, 0)}
	// 600/min = 10 tokens/sec; burst 5.
	l := newTestLimiter(RateLimit{RequestsPerMinute: 600, Burst: 5}, fc)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("Wait %d: %v", i, err)
		}
	}
	fc.now = fc.now.Add(200 * time.Millisecond) // exactly 2 tokens refill
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if len(fc.sleeps) != 0 {
		t.Fatalf("unexpected sleeps %v", fc.sleeps)
	}
	// Third needs another 100ms of refill.
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if len(fc.sleeps) != 1 || fc.sleeps[0] <= 0 {
		t.Fatalf("sleeps = %v", fc.sleeps)
	}
}

func TestRateLimiterBurstCapped(t *testing.T) {
	t.Parallel()
	fc := &fakeClock{now: time.Unix(1000, 0)}
	l := newTestLimiter(RateLimit{RequestsPerMinute: 600, Burst: 3}, fc)
	fc.now = fc.now.Add(time.Hour) // huge idle time
	for i := 0; i < 3; i++ {
		if err := l.Wait(context.Background()); err != nil {
			t.Fatalf("Wait %d: %v", i, err)
		}
	}
	// Fourth must wait even after a long idle: burst is capped at 3.
	if err := l.Wait(context.Background()); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if len(fc.sleeps) != 1 {
		t.Fatalf("sleeps = %v", fc.sleeps)
	}
}

func TestRateLimiterContextCancel(t *testing.T) {
	t.Parallel()
	fc := &fakeClock{now: time.Unix(1000, 0)}
	l := newTestLimiter(RateLimit{RequestsPerMinute: 1, Burst: 1}, fc)
	ctx, cancel := context.WithCancel(context.Background())
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	cancel()
	err := l.Wait(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestRateLimiterDisabled(t *testing.T) {
	t.Parallel()
	l := newRateLimiter(RateLimit{RequestsPerMinute: 0, Burst: 0}, nil)
	for i := 0; i < 100; i++ {
		if err := l.Wait(context.Background()); err != nil {
			t.Fatalf("Wait %d: %v", i, err)
		}
	}
	// A nil limiter is also safe.
	var nilLimiter *rateLimiter
	if err := nilLimiter.Wait(context.Background()); err != nil {
		t.Fatalf("nil limiter: %v", err)
	}
}

func TestRateLimiterNegativeRate(t *testing.T) {
	t.Parallel()
	l := newRateLimiter(RateLimit{RequestsPerMinute: -5, Burst: -1}, nil)
	if err := l.Wait(context.Background()); err != nil {
		t.Fatalf("Wait: %v", err)
	}
}

func TestSleepContext(t *testing.T) {
	t.Parallel()
	if err := sleepContext(context.Background(), time.Millisecond); err != nil {
		t.Fatalf("sleep: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleepContext(ctx, time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}
