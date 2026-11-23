//go:build integration

package geckoterminal_test

import (
	"context"
	"os"
	"testing"
	"time"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

// TestIntegration exercises the live public API. It is skipped unless
// GECKOTERMINAL_INTEGRATION=1 is set. The default client-side rate
// limiter (10 req/min) spaces out calls automatically, so the suite can
// take a while.
func TestIntegration(t *testing.T) {
	if os.Getenv("GECKOTERMINAL_INTEGRATION") == "" {
		t.Skip("set GECKOTERMINAL_INTEGRATION=1 to run")
	}
	client, err := geckoterminal.NewClient(geckoterminal.WithTimeout(30 * time.Second))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Minute)
	defer cancel()

	t.Run("networks", func(t *testing.T) {
		page, err := client.Networks.List(ctx, nil)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(page.Items) == 0 {
			t.Fatal("no networks")
		}
	})

	var poolAddress string
	t.Run("trending", func(t *testing.T) {
		page, err := client.Pools.Trending(ctx, nil)
		if err != nil {
			t.Fatalf("Trending: %v", err)
		}
		if len(page.Items) == 0 {
			t.Fatal("no trending pools")
		}
		for _, p := range page.Items {
			if p.Relationships.Network != nil && p.Relationships.Network.ID() == "eth" {
				poolAddress = p.Attributes.Address
				break
			}
		}
		if poolAddress == "" {
			poolAddress = page.Items[0].Attributes.Address
		}
	})

	t.Run("pool", func(t *testing.T) {
		if poolAddress == "" {
			t.Skip("no pool")
		}
		resp, err := client.Pools.Get(ctx, "eth", poolAddress, nil)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if resp.Data.Attributes.Name == "" {
			t.Fatal("empty pool name")
		}
	})

	t.Run("ohlcv", func(t *testing.T) {
		if poolAddress == "" {
			t.Skip("no pool")
		}
		resp, err := client.OHLCV.Get(ctx, "eth", poolAddress, geckoterminal.TimeframeDay, &geckoterminal.OHLCVOptions{
			Limit: geckoterminal.Ptr(10),
		})
		if err != nil {
			t.Fatalf("OHLCV: %v", err)
		}
		if len(resp.Data.Attributes.OHLCVList) == 0 {
			t.Fatal("no candles")
		}
	})

	t.Run("trades", func(t *testing.T) {
		if poolAddress == "" {
			t.Skip("no pool")
		}
		page, err := client.Trades.List(ctx, "eth", poolAddress, nil)
		if err != nil {
			t.Fatalf("Trades: %v", err)
		}
		t.Logf("got %d trades", len(page.Items))
	})

	t.Run("simple_price", func(t *testing.T) {
		resp, err := client.Simple.TokenPrices(ctx, "eth", []string{
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		}, nil)
		if err != nil {
			t.Fatalf("TokenPrices: %v", err)
		}
		if len(resp.Data.TokenPrices) == 0 {
			t.Fatal("no prices")
		}
	})
}
