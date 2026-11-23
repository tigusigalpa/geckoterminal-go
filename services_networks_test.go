package geckoterminal_test

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func TestNetworksList(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var rawQuery string
	c, _ := newTestClient(t, fixtureHandler(t, "networks.json", func(r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		rawQuery = r.URL.RawQuery
	}))
	page, err := c.Networks.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/networks" || gotMethod != http.MethodGet {
		t.Fatalf("%s %s", gotMethod, gotPath)
	}
	if rawQuery != "" {
		t.Fatalf("unexpected query %q", rawQuery)
	}
	if len(page.Items) != 3 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[0].ID != "eth" || page.Items[0].Attributes.Name != "Ethereum" {
		t.Fatalf("unexpected first item %+v", page.Items[0])
	}
	if page.Items[1].Attributes.CoinGeckoAssetPlatformID != "binance-smart-chain" {
		t.Fatalf("asset platform id = %q", page.Items[1].Attributes.CoinGeckoAssetPlatformID)
	}
	if page.Links == nil || page.Links.Next == nil {
		t.Fatal("links.next missing")
	}
	if !page.HasNext {
		t.Fatal("HasNext = false")
	}
	if page.Number != 1 {
		t.Fatalf("Number = %d", page.Number)
	}
}

func TestNetworksListWithPage(t *testing.T) {
	t.Parallel()
	var page string
	c, _ := newTestClient(t, fixtureHandler(t, "networks.json", func(r *http.Request) {
		page = r.URL.Query().Get("page")
	}))
	resp, err := c.Networks.List(context.Background(), &geckoterminal.ListOptions{Page: geckoterminal.Ptr(3)})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page != "3" {
		t.Fatalf("page query = %q", page)
	}
	if resp.Number != 3 {
		t.Fatalf("Number = %d", resp.Number)
	}
}

func TestNetworksListInvalidPage(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
	})
	for _, p := range []int{0, 11, -1} {
		_, err := c.Networks.List(context.Background(), &geckoterminal.ListOptions{Page: geckoterminal.Ptr(p)})
		var ve *geckoterminal.ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("page %d: err = %v, want *ValidationError", p, err)
		}
		if !errors.Is(err, geckoterminal.ErrInvalidArgument) {
			t.Fatalf("page %d: err does not unwrap to ErrInvalidArgument", p)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("server calls = %d, want 0", calls.Load())
	}
}

func TestNetworksListErrors(t *testing.T) {
	t.Parallel()
	t.Run("404", func(t *testing.T) {
		c, _ := newTestClient(t, statusHandler(http.StatusNotFound, string(fixture(t, "error_404.json")), nil))
		_, err := c.Networks.List(context.Background(), nil)
		if !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("429", func(t *testing.T) {
		c, _ := newTestClient(t, statusHandler(http.StatusTooManyRequests, string(fixture(t, "error_429.json")), nil))
		_, err := c.Networks.List(context.Background(), nil)
		if !geckoterminal.IsRateLimited(err) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestNetworksAll(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Write(fixture(t, "networks.json"))
	})
	var ids []string
	for n, err := range c.Networks.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iter err: %v", err)
		}
		ids = append(ids, n.ID)
	}
	// The fixture always reports links.next, so iteration runs the
	// full 10-page cap.
	if len(ids) != 3*10 {
		t.Fatalf("iterated %d items", len(ids))
	}
	if calls.Load() != 10 {
		t.Fatalf("calls = %d, want 10", calls.Load())
	}
}

func TestNetworksAllEarlyStop(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write(fixture(t, "networks.json"))
	})
	count := 0
	for _, err := range c.Networks.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iter err: %v", err)
		}
		count++
		if count == 1 {
			break
		}
	}
	if count != 1 || calls.Load() != 1 {
		t.Fatalf("count=%d calls=%d", count, calls.Load())
	}
}
