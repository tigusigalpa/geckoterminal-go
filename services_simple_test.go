package geckoterminal_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func TestSimpleTokenPrices(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "simple_price.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	resp, err := c.Simple.TokenPrices(context.Background(), "eth", []string{
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		"0xdac17f958d2ee523a2206206994597c13d831ec7",
	}, nil)
	if err != nil {
		t.Fatalf("TokenPrices: %v", err)
	}
	if gotPath != "/simple/networks/eth/token_price/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2,0xdac17f958d2ee523a2206206994597c13d831ec7" {
		t.Fatalf("path = %q", gotPath)
	}
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	if resp.Data.TokenPrices[weth] != "1722.41" {
		t.Fatalf("price = %q", resp.Data.TokenPrices[weth])
	}
	if resp.Data.MarketCapUSD[weth] != "3739865936.60058" {
		t.Fatalf("mcap = %q", resp.Data.MarketCapUSD[weth])
	}
	if resp.Data.H24VolumeUSD[weth] != "272228382.847295" {
		t.Fatalf("vol = %q", resp.Data.H24VolumeUSD[weth])
	}
	if resp.Data.H24PriceChangePercentage[weth] != "1.2954175267" {
		t.Fatalf("change = %q", resp.Data.H24PriceChangePercentage[weth])
	}
	if resp.Data.TotalReserveInUSD[weth] != "716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726" {
		t.Fatalf("reserve = %q", resp.Data.TotalReserveInUSD[weth])
	}
}

func TestSimpleTokenPricesOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "simple_price.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.Simple.TokenPrices(context.Background(), "eth", []string{"0xabc"}, &geckoterminal.TokenPricesOptions{
		IncludeMarketCap:         geckoterminal.Ptr(true),
		McapFDVFallback:          geckoterminal.Ptr(true),
		Include24hVol:            geckoterminal.Ptr(true),
		Include24hPriceChange:    geckoterminal.Ptr(true),
		IncludeTotalReserveInUSD: geckoterminal.Ptr(true),
		IncludeInactiveSource:    geckoterminal.Ptr(false),
	})
	if err != nil {
		t.Fatalf("TokenPrices: %v", err)
	}
	want := map[string]string{
		"include_market_cap":           "true",
		"mcap_fdv_fallback":            "true",
		"include_24hr_vol":             "true",
		"include_24hr_price_change":    "true",
		"include_total_reserve_in_usd": "true",
		"include_inactive_source":      "false",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Fatalf("%s = %q, want %q", k, q.Get(k), v)
		}
	}
}

func TestSimpleTokenPricesValidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	ctx := context.Background()
	if _, err := c.Simple.TokenPrices(ctx, "eth", nil, nil); err == nil {
		t.Fatal("empty addresses: expected error")
	}
	addrs := make([]string, 31)
	for i := range addrs {
		addrs[i] = "0xaaa"
	}
	_, err := c.Simple.TokenPrices(ctx, "eth", addrs, nil)
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("31 addresses: err = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}
