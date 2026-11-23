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

func TestOHLCVGet(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "ohlcv.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	resp, err := c.OHLCV.Get(context.Background(), "eth", "0xabc", geckoterminal.TimeframeDay, nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/networks/eth/pools/0xabc/ohlcv/day" {
		t.Fatalf("path = %q", gotPath)
	}
	candles := resp.Data.Attributes.OHLCVList
	if len(candles) != 3 {
		t.Fatalf("candles = %d", len(candles))
	}
	first := candles[0]
	if first.Timestamp.Unix() != 1712534400 {
		t.Fatalf("ts = %v", first.Timestamp)
	}
	if first.Open != "3454.61590249189" || first.Close != "3660.85954963415" {
		t.Fatalf("candle = %+v", first)
	}
	if resp.Meta.Base.Symbol != "WETH" || resp.Meta.Quote.Symbol != "USDT" {
		t.Fatalf("meta = %+v", resp.Meta)
	}
	if resp.Meta.Base.CoinGeckoCoinID == nil || *resp.Meta.Base.CoinGeckoCoinID != "weth" {
		t.Fatalf("base coin id = %v", resp.Meta.Base.CoinGeckoCoinID)
	}
}

func TestOHLCVGetAllOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "ohlcv.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.OHLCV.Get(context.Background(), "eth", "0xabc", geckoterminal.TimeframeMinute, &geckoterminal.OHLCVOptions{
		Aggregate:             geckoterminal.Ptr(geckoterminal.Aggregate15),
		BeforeTimestamp:       geckoterminal.Ptr(int64(1712534400)),
		Limit:                 geckoterminal.Ptr(1000),
		Currency:              geckoterminal.Ptr(geckoterminal.CurrencyToken),
		IncludeEmptyIntervals: geckoterminal.Ptr(true),
		Token:                 geckoterminal.Ptr("base"),
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := map[string]string{
		"aggregate":               "15",
		"before_timestamp":        "1712534400",
		"limit":                   "1000",
		"currency":                "token",
		"include_empty_intervals": "true",
		"token":                   "base",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Fatalf("%s = %q, want %q", k, q.Get(k), v)
		}
	}
}

func TestOHLCVAggregateValidation(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, fixtureHandler(t, "ohlcv.json", nil))
	ctx := context.Background()
	valid := []struct {
		tf  geckoterminal.Timeframe
		agg geckoterminal.Aggregate
	}{
		{geckoterminal.TimeframeDay, geckoterminal.Aggregate1},
		{geckoterminal.TimeframeHour, geckoterminal.Aggregate1},
		{geckoterminal.TimeframeHour, geckoterminal.Aggregate4},
		{geckoterminal.TimeframeHour, geckoterminal.Aggregate12},
		{geckoterminal.TimeframeMinute, geckoterminal.Aggregate1},
		{geckoterminal.TimeframeMinute, geckoterminal.Aggregate5},
		{geckoterminal.TimeframeMinute, geckoterminal.Aggregate15},
	}
	for _, tc := range valid {
		if _, err := c.OHLCV.Get(ctx, "eth", "0xabc", tc.tf, &geckoterminal.OHLCVOptions{
			Aggregate: geckoterminal.Ptr(tc.agg),
		}); err != nil {
			t.Fatalf("%s/%s: %v", tc.tf, tc.agg, err)
		}
	}
	invalid := []struct {
		tf  geckoterminal.Timeframe
		agg geckoterminal.Aggregate
	}{
		{geckoterminal.TimeframeDay, geckoterminal.Aggregate4},
		{geckoterminal.TimeframeHour, geckoterminal.Aggregate5},
		{geckoterminal.TimeframeMinute, geckoterminal.Aggregate12},
	}
	for _, tc := range invalid {
		_, err := c.OHLCV.Get(ctx, "eth", "0xabc", tc.tf, &geckoterminal.OHLCVOptions{
			Aggregate: geckoterminal.Ptr(tc.agg),
		})
		var ve *geckoterminal.ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("%s/%s: err = %v", tc.tf, tc.agg, err)
		}
	}
}

func TestOHLCVValidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	ctx := context.Background()
	var ve *geckoterminal.ValidationError

	_, err := c.OHLCV.Get(ctx, "eth", "0xabc", geckoterminal.Timeframe("week"), nil)
	if !errors.As(err, &ve) {
		t.Fatalf("timeframe: err = %v", err)
	}
	_, err = c.OHLCV.Get(ctx, "eth", "0xabc", geckoterminal.TimeframeDay, &geckoterminal.OHLCVOptions{
		Limit: geckoterminal.Ptr(1001),
	})
	if !errors.As(err, &ve) {
		t.Fatalf("limit=1001: err = %v", err)
	}
	_, err = c.OHLCV.Get(ctx, "eth", "0xabc", geckoterminal.TimeframeDay, &geckoterminal.OHLCVOptions{
		Limit: geckoterminal.Ptr(0),
	})
	if !errors.As(err, &ve) {
		t.Fatalf("limit=0: err = %v", err)
	}
	_, err = c.OHLCV.Get(ctx, "eth", "0xabc", geckoterminal.TimeframeDay, &geckoterminal.OHLCVOptions{
		BeforeTimestamp: geckoterminal.Ptr(int64(-5)),
	})
	if !errors.As(err, &ve) {
		t.Fatalf("before_timestamp: err = %v", err)
	}
	_, err = c.OHLCV.Get(ctx, "eth", "0xabc", geckoterminal.TimeframeDay, &geckoterminal.OHLCVOptions{
		Currency: geckoterminal.Ptr(geckoterminal.Currency("eur")),
	})
	if !errors.As(err, &ve) {
		t.Fatalf("currency: err = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestOHLCVCandles(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, fixtureHandler(t, "ohlcv.json", nil))
	candles, err := c.OHLCV.Candles(context.Background(), "eth", "0xabc", geckoterminal.TimeframeHour, nil)
	if err != nil {
		t.Fatalf("Candles: %v", err)
	}
	if len(candles) != 3 {
		t.Fatalf("candles = %d", len(candles))
	}
}
