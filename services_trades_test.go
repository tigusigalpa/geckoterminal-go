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

func TestTradesList(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "trades.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	page, err := c.Trades.List(context.Background(), "eth", "0x60594a405d53811d3bc4766596efd80fd545a270", nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/networks/eth/pools/0x60594a405d53811d3bc4766596efd80fd545a270/trades" {
		t.Fatalf("path = %q", gotPath)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
	tr := page.Items[0].Attributes
	if tr.Kind != geckoterminal.TradeKindSell {
		t.Fatalf("kind = %q", tr.Kind)
	}
	if tr.BlockNumber != 26475103 {
		t.Fatalf("block = %d", tr.BlockNumber)
	}
	if tr.BlockTimestamp.IsZero() {
		t.Fatal("block_timestamp missing")
	}
	if tr.FromTokenAmount != "1846.92134598154" {
		t.Fatalf("from amount = %q", tr.FromTokenAmount)
	}
	if tr.VolumeInUSD != "1832.68169910395" {
		t.Fatalf("volume = %q", tr.VolumeInUSD)
	}
	if page.Items[1].Attributes.Kind != geckoterminal.TradeKindBuy || !page.Items[1].Attributes.IsAnomalous {
		t.Fatalf("second trade = %+v", page.Items[1].Attributes)
	}
	if page.HasNext {
		t.Fatal("HasNext should be false without links or cursor")
	}
}

func TestTradesListAllOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "trades.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.Trades.List(context.Background(), "eth", "0xabc", &geckoterminal.TradesOptions{
		TradeVolumeInUSDGreaterThan: geckoterminal.Ptr(geckoterminal.Decimal("100.5")),
		IncludeAnomalous:            geckoterminal.Ptr(true),
		TradingPeriod:               geckoterminal.Ptr(geckoterminal.TradingPeriod7d),
		Cursor:                      geckoterminal.Ptr("cur123"),
		PerPage:                     geckoterminal.Ptr(300),
		Token:                       geckoterminal.Ptr("quote"),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := map[string]string{
		"trade_volume_in_usd_greater_than": "100.5",
		"include_anomalous":                "true",
		"trading_period":                   "7d",
		"cursor":                           "cur123",
		"per_page":                         "300",
		"token":                            "quote",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Fatalf("%s = %q, want %q", k, q.Get(k), v)
		}
	}
}

func TestTradesListValidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	ctx := context.Background()
	var ve *geckoterminal.ValidationError

	_, err := c.Trades.List(ctx, "eth", "0xabc", &geckoterminal.TradesOptions{PerPage: geckoterminal.Ptr(301)})
	if !errors.As(err, &ve) {
		t.Fatalf("per_page=301: err = %v", err)
	}
	_, err = c.Trades.List(ctx, "eth", "0xabc", &geckoterminal.TradesOptions{PerPage: geckoterminal.Ptr(0)})
	if !errors.As(err, &ve) {
		t.Fatalf("per_page=0: err = %v", err)
	}
	_, err = c.Trades.List(ctx, "eth", "0xabc", &geckoterminal.TradesOptions{
		TradingPeriod: geckoterminal.Ptr(geckoterminal.TradingPeriod("1h")),
	})
	if !errors.As(err, &ve) {
		t.Fatalf("trading_period: err = %v", err)
	}
	_, err = c.Trades.List(ctx, "eth", "0xabc", &geckoterminal.TradesOptions{
		Token: geckoterminal.Ptr("bad token"),
	})
	if !errors.As(err, &ve) {
		t.Fatalf("token: err = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestTradesListNextCursor(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"data":[],"meta":{"next_cursor":"abc123"}}`))
	})
	page, err := c.Trades.List(context.Background(), "eth", "0xabc", &geckoterminal.TradesOptions{
		Cursor: geckoterminal.Ptr("xyz"),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if page.NextCursor() != "abc123" {
		t.Fatalf("NextCursor = %q", page.NextCursor())
	}
	if !page.HasNext {
		t.Fatal("HasNext should be true with next_cursor")
	}
}
