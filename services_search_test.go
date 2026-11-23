package geckoterminal_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func TestSearchPools(t *testing.T) {
	t.Parallel()
	var gotPath, query string
	c, _ := newTestClient(t, fixtureHandler(t, "search_pools.json", func(r *http.Request) {
		gotPath = r.URL.Path
		query = r.URL.Query().Get("query")
	}))
	page, err := c.Search.Pools(context.Background(), "weth", nil)
	if err != nil {
		t.Fatalf("Pools: %v", err)
	}
	if gotPath != "/search/pools" {
		t.Fatalf("path = %q", gotPath)
	}
	if query != "weth" {
		t.Fatalf("query = %q", query)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[0].Attributes.Name != "WETH / USDC 0.05%" {
		t.Fatalf("name = %q", page.Items[0].Attributes.Name)
	}
}

func TestSearchPoolsOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "search_pools.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.Search.Pools(context.Background(), "weth", &geckoterminal.SearchPoolsOptions{
		Network: geckoterminal.Ptr("eth"),
		Include: []geckoterminal.Include{geckoterminal.IncludeDEX},
		Page:    geckoterminal.Ptr(2),
	})
	if err != nil {
		t.Fatalf("Pools: %v", err)
	}
	if q.Get("network") != "eth" || q.Get("include") != "dex" || q.Get("page") != "2" {
		t.Fatalf("q = %v", q)
	}
}

func TestSearchPoolsValidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	ctx := context.Background()
	var ve *geckoterminal.ValidationError

	if _, err := c.Search.Pools(ctx, "", nil); !errors.As(err, &ve) {
		t.Fatalf("empty query: err = %v", err)
	}
	if _, err := c.Search.Pools(ctx, strings.Repeat("x", 257), nil); !errors.As(err, &ve) {
		t.Fatalf("long query: err = %v", err)
	}
	bad := "bad net"
	if _, err := c.Search.Pools(ctx, "weth", &geckoterminal.SearchPoolsOptions{Network: &bad}); !errors.As(err, &ve) {
		t.Fatalf("bad network: err = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}
