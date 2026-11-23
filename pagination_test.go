package geckoterminal_test

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func TestAllStopsWhenHasNextFalse(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Write([]byte(`{"data":[{"id":"x","type":"network","attributes":{"name":"n"}}]}`))
			return
		}
		w.Write([]byte(`{"data":[]}`))
	})
	count := 0
	for _, err := range c.Networks.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iter: %v", err)
		}
		count++
	}
	// One non-empty page still reports HasNext via the heuristic
	// (no links), so a second empty page is fetched before stopping.
	if count != 1 || calls.Load() != 2 {
		t.Fatalf("count=%d calls=%d", count, calls.Load())
	}
}

func TestAllPropagatesError(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Write([]byte(`{"data":[{"id":"x","type":"network","attributes":{"name":"n"}}]}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`boom`))
	})
	var got error
	count := 0
	for _, err := range c.Networks.All(context.Background(), nil) {
		if err != nil {
			got = err
			break
		}
		count++
	}
	if got == nil {
		t.Fatal("expected error")
	}
	var apiErr *geckoterminal.APIError
	if !errors.As(got, &apiErr) || apiErr.StatusCode != 500 {
		t.Fatalf("err = %v", got)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1 item before the error", count)
	}
}

func TestAllContextCancel(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"data":[{"id":"x","type":"network","attributes":{"name":"n"}}],"links":{"next":"http://x?page=2"}}`))
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	count := 0
	var got error
	for _, err := range c.Networks.All(ctx, nil) {
		if err != nil {
			got = err
			break
		}
		count++
		cancel()
	}
	if got == nil || !errors.Is(got, context.Canceled) {
		t.Fatalf("err = %v", got)
	}
	if count != 1 {
		t.Fatalf("count = %d", count)
	}
}

func TestAllNilPage(t *testing.T) {
	t.Parallel()
	fetch := func(_ context.Context, _ int) (*geckoterminal.Page[int], error) {
		return nil, nil
	}
	count := 0
	for _, err := range geckoterminal.All(context.Background(), fetch) {
		if err != nil {
			t.Fatalf("iter: %v", err)
		}
		count++
	}
	if count != 0 {
		t.Fatalf("count = %d", count)
	}
}

func TestCollect(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, fixtureHandler(t, "networks.json", nil))
	// networks.json advertises links.next forever, so Collect walks
	// the full 10-page cap.
	items, err := geckoterminal.Collect(c.Networks.All(context.Background(), nil))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(items) != 30 {
		t.Fatalf("items = %d", len(items))
	}
}

func TestCollectError(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, statusHandler(http.StatusInternalServerError, "boom", nil))
	items, err := geckoterminal.Collect(c.Networks.All(context.Background(), nil))
	if err == nil {
		t.Fatal("expected error")
	}
	if len(items) != 0 {
		t.Fatalf("items = %d", len(items))
	}
}
