package geckoterminal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryCacheGetSet(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache(2)
	entry := CacheEntry{Body: []byte("body"), Header: http.Header{"X": {"1"}}, ETag: `"tag"`}
	c.Set("k", entry, time.Minute)
	got, ok := c.Get("k")
	if !ok {
		t.Fatal("Get: not found")
	}
	if string(got.Body) != "body" || got.ETag != `"tag"` {
		t.Fatalf("entry = %+v", got)
	}
	if !got.fresh(time.Now()) {
		t.Fatal("entry not fresh")
	}
	if _, ok := c.Get("missing"); ok {
		t.Fatal("unexpected hit")
	}
}

func TestMemoryCacheExpiry(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache(4)
	c.now = func() time.Time { return time.Unix(1000, 0) }
	c.Set("k", CacheEntry{Body: []byte("b")}, time.Second)
	got, ok := c.Get("k")
	if !ok {
		t.Fatal("Get: not found")
	}
	if got.fresh(time.Unix(1002, 0)) {
		t.Fatal("entry should be expired")
	}
	// Expired entries are still returned for revalidation.
	if got.ExpiresAt != time.Unix(1001, 0) {
		t.Fatalf("ExpiresAt = %v", got.ExpiresAt)
	}
}

func TestMemoryCacheEvictsOldest(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache(2)
	c.Set("a", CacheEntry{}, time.Minute)
	c.Set("b", CacheEntry{}, time.Minute)
	c.Get("a") // promote a
	c.Set("c", CacheEntry{}, time.Minute)
	if _, ok := c.Get("b"); ok {
		t.Fatal("b should have been evicted")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("a should survive")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("c should survive")
	}
}

func TestMemoryCacheZeroTTL(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache(1)
	c.Set("k", CacheEntry{}, 0)
	if _, ok := c.Get("k"); ok {
		t.Fatal("zero-TTL entry stored")
	}
}

func TestCacheTTLParsing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cc   string
		want time.Duration
	}{
		{"empty", "", defaultCacheTTL},
		{"max-age", "max-age=60", time.Minute},
		{"s-maxage wins", "max-age=60, s-maxage=30", 30 * time.Second},
		{"real api", "max-age=30, public, must-revalidate, s-maxage=60", time.Minute},
		{"no-store", "no-store", -1},
		{"no-cache", "private, no-cache", -1},
		{"private", "private, max-age=60", -1},
		{"garbage", "max-age=abc", defaultCacheTTL},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{"Cache-Control": {tc.cc}}
			if got := cacheTTL(h, defaultCacheTTL); got != tc.want {
				t.Fatalf("cacheTTL(%q) = %v, want %v", tc.cc, got, tc.want)
			}
		})
	}
}

func TestClientCacheHit(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Cache-Control", "max-age=60")
		_, _ = w.Write([]byte(`{"data":[{"id":"eth","type":"network","attributes":{"name":"Ethereum"}}]}`))
	}))
	defer srv.Close()
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithRateLimit(0, 0),
		WithRetryPolicy(RetryPolicy{MaxRetries: 0}),
		WithCache(NewMemoryCache(8)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	for i := 0; i < 2; i++ {
		page, err := c.Networks.List(context.Background(), nil)
		if err != nil {
			t.Fatalf("List %d: %v", i, err)
		}
		if len(page.Items) != 1 {
			t.Fatalf("items = %d", len(page.Items))
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("server calls = %d, want 1 (second served from cache)", calls.Load())
	}
}

func TestClientCacheETagRevalidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 2 {
			if r.Header.Get("If-None-Match") != `"v1"` {
				t.Errorf("If-None-Match = %q", r.Header.Get("If-None-Match"))
			}
			w.Header().Set("Cache-Control", "max-age=0")
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		w.Header().Set("Cache-Control", "max-age=60")
		_, _ = w.Write([]byte(`{"data":[{"id":"eth","type":"network","attributes":{"name":"Ethereum"}}]}`))
	}))
	defer srv.Close()
	cache := NewMemoryCache(8)
	cache.now = func() time.Time {
		// Entries are instantly stale but keep their ETag.
		return time.Now().Add(-time.Hour)
	}
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithRateLimit(0, 0),
		WithRetryPolicy(RetryPolicy{MaxRetries: 0}),
		WithCache(cache),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	for i := 0; i < 2; i++ {
		page, err := c.Networks.List(context.Background(), nil)
		if err != nil {
			t.Fatalf("List %d: %v", i, err)
		}
		if len(page.Items) != 1 || page.Items[0].ID != "eth" {
			t.Fatalf("items = %+v", page.Items)
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (conditional request)", calls.Load())
	}
}

func TestClientCacheNoStore(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithRateLimit(0, 0),
		WithCache(NewMemoryCache(8)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := c.Networks.List(context.Background(), nil); err != nil {
			t.Fatalf("List: %v", err)
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (no-store not cached)", calls.Load())
	}
}
