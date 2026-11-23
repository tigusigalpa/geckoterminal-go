package geckoterminal_test

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func TestDexesListByNetwork(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "dexes.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	page, err := c.Dexes.ListByNetwork(context.Background(), "eth", nil)
	if err != nil {
		t.Fatalf("ListByNetwork: %v", err)
	}
	if gotPath != "/networks/eth/dexes" {
		t.Fatalf("path = %q", gotPath)
	}
	if len(page.Items) != 3 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[0].ID != "uniswap_v3" || page.Items[0].Attributes.Name != "Uniswap V3" {
		t.Fatalf("unexpected item %+v", page.Items[0])
	}
}

func TestDexesListByNetworkPage(t *testing.T) {
	t.Parallel()
	var page string
	c, _ := newTestClient(t, fixtureHandler(t, "dexes.json", func(r *http.Request) {
		page = r.URL.Query().Get("page")
	}))
	if _, err := c.Dexes.ListByNetwork(context.Background(), "eth", &geckoterminal.DEXListOptions{Page: geckoterminal.Ptr(2)}); err != nil {
		t.Fatalf("ListByNetwork: %v", err)
	}
	if page != "2" {
		t.Fatalf("page = %q", page)
	}
}

func TestDexesListByNetworkValidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	for _, network := range []string{"", "et/h", "a b", "net?x"} {
		_, err := c.Dexes.ListByNetwork(context.Background(), network, nil)
		var ve *geckoterminal.ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("network %q: err = %v", network, err)
		}
	}
	_, err := c.Dexes.ListByNetwork(context.Background(), "eth", &geckoterminal.DEXListOptions{Page: geckoterminal.Ptr(11)})
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("page: err = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestDexesListByNetwork404(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, statusHandler(http.StatusNotFound, string(fixture(t, "error_404.json")), nil))
	_, err := c.Dexes.ListByNetwork(context.Background(), "nope", nil)
	if !geckoterminal.IsNotFound(err) {
		t.Fatalf("err = %v", err)
	}
}

func TestDexesAllByNetwork(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{"data":[]}`))
	})
	count := 0
	for _, err := range c.Dexes.AllByNetwork(context.Background(), "eth", nil) {
		if err != nil {
			t.Fatalf("iter: %v", err)
		}
		count++
	}
	if count != 0 {
		t.Fatalf("count = %d", count)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1 (empty page stops iteration)", calls.Load())
	}
}
