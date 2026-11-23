package geckoterminal_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func TestEnumStringers(t *testing.T) {
	t.Parallel()
	if geckoterminal.Duration5m.String() != "5m" || !geckoterminal.Duration24h.Valid() ||
		geckoterminal.Duration("2h").Valid() {
		t.Fatal("Duration")
	}
	if geckoterminal.SortH24VolumeUSDDesc.String() != "h24_volume_usd_desc" {
		t.Fatal("Sort")
	}
	if geckoterminal.IncludeBaseToken.String() != "base_token" {
		t.Fatal("Include")
	}
	if geckoterminal.TradingPeriod30d.String() != "30d" {
		t.Fatal("TradingPeriod")
	}
	if geckoterminal.TimeframeHour.String() != "hour" {
		t.Fatal("Timeframe")
	}
	if geckoterminal.Aggregate15.String() != "15" {
		t.Fatal("Aggregate")
	}
	if geckoterminal.CurrencyUSD.String() != "usd" {
		t.Fatal("Currency")
	}
	if geckoterminal.IntervalH24.String() != "h24" {
		t.Fatal("Interval")
	}
	if geckoterminal.TradeKindBuy.String() != "buy" {
		t.Fatal("TradeKind")
	}
	if !geckoterminal.TimeframeDay.Valid() || geckoterminal.Timeframe("x").Valid() {
		t.Fatal("Timeframe.Valid")
	}
}

func TestIntervalHelpers(t *testing.T) {
	t.Parallel()
	v := geckoterminal.IntervalValues{
		geckoterminal.IntervalM5:  "1",
		geckoterminal.IntervalM15: "2",
		geckoterminal.IntervalM30: "3",
		geckoterminal.IntervalH1:  "4",
		geckoterminal.IntervalH6:  "5",
		geckoterminal.IntervalH24: "6",
	}
	if v.M5() != "1" || v.M15() != "2" || v.M30() != "3" ||
		v.H1() != "4" || v.H6() != "5" || v.H24() != "6" ||
		v.Get(geckoterminal.Interval("m1")) != "" {
		t.Fatal("IntervalValues helpers")
	}
	tx := geckoterminal.Transactions{
		geckoterminal.IntervalH24: {Buys: 24},
		geckoterminal.IntervalH6:  {Buys: 6},
		geckoterminal.IntervalH1:  {Buys: 1},
		geckoterminal.IntervalM5:  {Buys: 5},
	}
	if tx.H24().Buys != 24 || tx.H6().Buys != 6 || tx.H1().Buys != 1 ||
		tx.M5().Buys != 5 || tx.Get(geckoterminal.IntervalM15).Buys != 0 {
		t.Fatal("Transactions helpers")
	}
}

func TestRelationshipEdgeCases(t *testing.T) {
	t.Parallel()
	var r geckoterminal.Relationship
	if err := json.Unmarshal([]byte(`{"data":[{"id":"a","type":"x"},{"id":"b","type":"x"}]}`), &r); err != nil {
		t.Fatal(err)
	}
	if r.ID() != "" {
		t.Fatal("ID on array data should be empty")
	}
	if ids := r.IDs(); len(ids) != 2 || ids[0] != "a" {
		t.Fatalf("IDs = %v", ids)
	}
	var single geckoterminal.Relationship
	if err := json.Unmarshal([]byte(`{"data":{"id":"a","type":"x"}}`), &single); err != nil {
		t.Fatal(err)
	}
	if single.IDs() != nil {
		t.Fatal("IDs on object data should be nil")
	}
	var empty geckoterminal.Relationship
	if empty.ID() != "" || empty.IDs() != nil {
		t.Fatal("empty relationship")
	}
}

func TestIncludedLookupMiss(t *testing.T) {
	t.Parallel()
	var inc geckoterminal.Included
	if err := json.Unmarshal([]byte(`[{"id":"x","type":"token","attributes":{"name":"t","address":"a","symbol":"s","decimals":1,"image_url":"u"}}]`), &inc); err != nil {
		t.Fatal(err)
	}
	if _, ok := inc.Token("missing"); ok {
		t.Fatal("Token miss")
	}
	if _, ok := inc.Pool("missing"); ok {
		t.Fatal("Pool miss")
	}
	if _, ok := inc.DEX("missing"); ok {
		t.Fatal("DEX miss")
	}
	if _, ok := inc.Network("missing"); ok {
		t.Fatal("Network miss")
	}
}

func TestResponseDecodeErrors(t *testing.T) {
	t.Parallel()
	var r geckoterminal.Response[geckoterminal.Pool]
	if err := json.Unmarshal([]byte(`{}`), &r); err == nil {
		t.Fatal("missing data should error")
	}
	if err := json.Unmarshal([]byte(`{"data":null}`), &r); err == nil {
		t.Fatal("null data should error")
	}
	if err := json.Unmarshal([]byte(`{"data":"x"}`), &r); err == nil {
		t.Fatal("wrong data type should error")
	}
	if err := json.Unmarshal([]byte(`{bad`), &r); err == nil {
		t.Fatal("malformed json should error")
	}
	var p geckoterminal.Page[geckoterminal.Pool]
	if err := json.Unmarshal([]byte(`{bad`), &p); err == nil {
		t.Fatal("malformed page should error")
	}
	if err := json.Unmarshal([]byte(`{"data":{}}`), &p); err == nil {
		t.Fatal("object data should error")
	}
}

func TestPageNextCursorEdgeCases(t *testing.T) {
	t.Parallel()
	var p geckoterminal.Page[int]
	if p.NextCursor() != "" {
		t.Fatal("nil meta cursor")
	}
	p.Meta = json.RawMessage(`123`)
	if p.NextCursor() != "" {
		t.Fatal("bad meta cursor")
	}
	p.Meta = json.RawMessage(`{"next_cursor":"c1"}`)
	if p.NextCursor() != "c1" {
		t.Fatal("cursor")
	}
}

func TestAllTrending(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Write(fixture(t, "pools_trending.json"))
			return
		}
		w.Write([]byte(`{"data":[]}`))
	})
	items, err := geckoterminal.Collect(c.Pools.AllTrending(context.Background(), &geckoterminal.TrendingPoolsOptions{
		Duration: geckoterminal.Ptr(geckoterminal.Duration24h),
	}))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(items) != 2 || calls.Load() != 2 {
		t.Fatalf("items=%d calls=%d", len(items), calls.Load())
	}
}

func TestAllIteratorsWithOptions(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[]}`))
	})
	ctx := context.Background()
	for _, err := range c.Networks.All(ctx, &geckoterminal.ListOptions{Page: geckoterminal.Ptr(1)}) {
		if err != nil {
			t.Fatalf("networks: %v", err)
		}
	}
	for _, err := range c.Dexes.AllByNetwork(ctx, "eth", &geckoterminal.DEXListOptions{Page: geckoterminal.Ptr(1)}) {
		if err != nil {
			t.Fatalf("dexes: %v", err)
		}
	}
	for _, err := range c.Tokens.AllRecentlyUpdated(ctx, &geckoterminal.RecentlyUpdatedOptions{
		Network: geckoterminal.Ptr("eth"),
	}) {
		if err != nil {
			t.Fatalf("tokens: %v", err)
		}
	}
}

func TestServiceErrorPaths(t *testing.T) {
	t.Parallel()
	notFound := statusHandler(http.StatusNotFound, string(fixture(t, "error_404.json")), nil)
	ctx := context.Background()
	mk := func(t *testing.T) *geckoterminal.Client {
		c, _ := newTestClient(t, notFound)
		return c
	}
	t.Run("pools multi", func(t *testing.T) {
		if _, err := mk(t).Pools.Multi(ctx, "eth", []string{"0xabc"}, nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("tokens multi", func(t *testing.T) {
		if _, err := mk(t).Tokens.Multi(ctx, "eth", []string{"0xabc"}, nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("ohlcv", func(t *testing.T) {
		if _, err := mk(t).OHLCV.Get(ctx, "eth", "0xabc", geckoterminal.TimeframeDay, nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("ohlcv candles", func(t *testing.T) {
		if _, err := mk(t).OHLCV.Candles(ctx, "eth", "0xabc", geckoterminal.TimeframeDay, nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("simple", func(t *testing.T) {
		if _, err := mk(t).Simple.TokenPrices(ctx, "eth", []string{"0xabc"}, nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("trades", func(t *testing.T) {
		if _, err := mk(t).Trades.List(ctx, "eth", "0xabc", nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("search", func(t *testing.T) {
		if _, err := mk(t).Search.Pools(ctx, "weth", nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("tokens info", func(t *testing.T) {
		if _, err := mk(t).Tokens.Info(ctx, "eth", "0xabc"); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("tokens info validation", func(t *testing.T) {
		if _, err := mk(t).Tokens.Info(ctx, "eth", ""); err == nil {
			t.Fatal("expected validation error")
		}
	})
	t.Run("pool tokens info", func(t *testing.T) {
		if _, err := mk(t).Pools.TokensInfo(ctx, "eth", "0xabc", nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("new by network", func(t *testing.T) {
		if _, err := mk(t).Pools.NewByNetwork(ctx, "eth", nil); !geckoterminal.IsNotFound(err) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestMalformedResponses(t *testing.T) {
	t.Parallel()
	bad := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{oops`)) }
	ctx := context.Background()
	t.Run("pools get", func(t *testing.T) {
		c, _ := newTestClient(t, bad)
		if _, err := c.Pools.Get(ctx, "eth", "0xabc", nil); err == nil {
			t.Fatal("expected decode error")
		}
	})
	t.Run("pools multi", func(t *testing.T) {
		c, _ := newTestClient(t, bad)
		if _, err := c.Pools.Multi(ctx, "eth", []string{"0xabc"}, nil); err == nil {
			t.Fatal("expected decode error")
		}
	})
	t.Run("ohlcv", func(t *testing.T) {
		c, _ := newTestClient(t, bad)
		if _, err := c.OHLCV.Get(ctx, "eth", "0xabc", geckoterminal.TimeframeDay, nil); err == nil {
			t.Fatal("expected decode error")
		}
	})
	t.Run("simple", func(t *testing.T) {
		c, _ := newTestClient(t, bad)
		if _, err := c.Simple.TokenPrices(ctx, "eth", []string{"0xabc"}, nil); err == nil {
			t.Fatal("expected decode error")
		}
	})
	t.Run("tokens multi", func(t *testing.T) {
		c, _ := newTestClient(t, bad)
		if _, err := c.Tokens.Multi(ctx, "eth", []string{"0xabc"}, nil); err == nil {
			t.Fatal("expected decode error")
		}
	})
}

func TestPoolsMultiOptionsValidation(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	_, err := c.Pools.Multi(context.Background(), "eth", []string{"0xabc"}, &geckoterminal.MultiPoolOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeNetwork},
	})
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
}

func TestTokensMultiOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "tokens_multi.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.Tokens.Multi(context.Background(), "eth", []string{"0xabc"}, &geckoterminal.TokenMultiOptions{
		Include:               []geckoterminal.Include{geckoterminal.IncludeTopPools},
		IncludeComposition:    geckoterminal.Ptr(true),
		IncludeInactiveSource: geckoterminal.Ptr(true),
	})
	if err != nil {
		t.Fatalf("Multi: %v", err)
	}
	if q.Get("include") != "top_pools" || q.Get("include_composition") != "true" || q.Get("include_inactive_source") != "true" {
		t.Fatalf("q = %v", q)
	}
}

func TestTokensMultiOptionsValidation(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	_, err := c.Tokens.Multi(context.Background(), "eth", []string{"0xabc"}, &geckoterminal.TokenMultiOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeDEX},
	})
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
}

func TestPoolQueryValidationCoverage(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	ctx := context.Background()
	var ve *geckoterminal.ValidationError
	cases := make([]error, 0, 18)
	_, e1 := c.Pools.TopByDEX(ctx, "eth", "bad/dex", nil)
	cases = append(cases, e1)
	_, e2 := c.Pools.TopByDEX(ctx, "eth", "dex", &geckoterminal.PoolListOptions{Page: geckoterminal.Ptr(0)})
	cases = append(cases, e2)
	_, e3 := c.Pools.TopByToken(ctx, "eth", "0xabc", &geckoterminal.TokenPoolsOptions{Page: geckoterminal.Ptr(99)})
	cases = append(cases, e3)
	_, e4 := c.Pools.TopByToken(ctx, "eth", "0xabc", &geckoterminal.TokenPoolsOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeTopPools},
	})
	cases = append(cases, e4)
	_, e5 := c.Pools.TopByToken(ctx, "", "0xabc", nil)
	cases = append(cases, e5)
	_, e6 := c.Pools.TrendingByNetwork(ctx, "", nil)
	cases = append(cases, e6)
	_, e7 := c.Pools.NewByNetwork(ctx, "", nil)
	cases = append(cases, e7)
	_, e8 := c.Pools.New(ctx, &geckoterminal.NewPoolsOptions{Page: geckoterminal.Ptr(11)})
	cases = append(cases, e8)
	_, e9 := c.Pools.Get(ctx, "", "0xabc", nil)
	cases = append(cases, e9)
	_, e10 := c.Pools.Multi(ctx, "", []string{"0xabc"}, nil)
	cases = append(cases, e10)
	_, e11 := c.Pools.TokensInfo(ctx, "eth", "", nil)
	cases = append(cases, e11)
	_, e12 := c.Tokens.Get(ctx, "", "0xabc", nil)
	cases = append(cases, e12)
	_, e13 := c.Tokens.Multi(ctx, "", []string{"0xabc"}, nil)
	cases = append(cases, e13)
	_, e14 := c.Trades.List(ctx, "", "0xabc", nil)
	cases = append(cases, e14)
	_, e15 := c.OHLCV.Get(ctx, "", "0xabc", geckoterminal.TimeframeDay, nil)
	cases = append(cases, e15)
	_, e16 := c.Simple.TokenPrices(ctx, "", []string{"0xabc"}, nil)
	cases = append(cases, e16)
	_, e17 := c.Pools.Trending(ctx, &geckoterminal.TrendingPoolsOptions{Page: geckoterminal.Ptr(-2)})
	cases = append(cases, e17)
	_, e18 := c.Pools.TopByToken(ctx, "eth", "bad addr", nil)
	cases = append(cases, e18)
	for i, err := range cases {
		if !errors.As(err, &ve) {
			t.Fatalf("case %d: err = %v", i, err)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestClientTransportError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // all requests now fail at the transport level
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(url),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
			MaxRetries: 1,
			BaseDelay:  time.Millisecond,
			MaxDelay:   5 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Networks.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected transport error")
	}
	var apiErr *geckoterminal.APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("transport error must not be APIError: %v", err)
	}
}

func TestClientLoggerOption(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithLogger(slog.Default()),
		geckoterminal.WithClock(time.Now),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestClientAPIKeyOnly(t *testing.T) {
	t.Parallel()
	var key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key = r.Header.Get("x-cg-pro-api-key")
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithAPIKey("k-1"),
		geckoterminal.WithRateLimit(0, 0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if key != "k-1" {
		t.Fatalf("key = %q", key)
	}
}

func TestAPIErrorTitleFallback(t *testing.T) {
	t.Parallel()
	e := &geckoterminal.APIError{StatusCode: 599, Method: "GET", Path: "/x"}
	if got := e.Error(); !contains(got, "599") {
		t.Fatalf("Error() = %q", got)
	}
}

func TestMustNewClientSuccess(t *testing.T) {
	t.Parallel()
	if c := geckoterminal.MustNewClient(); c == nil {
		t.Fatal("nil client")
	}
}

func TestWithBaseURLParseError(t *testing.T) {
	t.Parallel()
	if _, err := geckoterminal.NewClient(geckoterminal.WithBaseURL("ht\ttp://bad")); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestClientOversizedBody(t *testing.T) {
	t.Parallel()
	big := make([]byte, (32<<20)+1)
	for i := range big {
		big[i] = 'x'
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(big)
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Networks.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected size limit error")
	}
	var apiErr *geckoterminal.APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("size error must not be APIError: %v", err)
	}
}

func TestClientRetryContextCancelDuringSleep(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`boom`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
			MaxRetries: 5,
			BaseDelay:  time.Hour, // sleep long enough to observe cancel
			MaxDelay:   time.Hour,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = c.Networks.List(ctx, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1 (no retry after cancel)", calls.Load())
	}
}

func TestClientTransportErrorRetryCancel(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(url),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
			MaxRetries: 5,
			BaseDelay:  time.Hour,
			MaxDelay:   time.Hour,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = c.Networks.List(ctx, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
}

func TestAllTopByNetworkWithOptions(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Write(fixture(t, "pools_trending.json"))
			return
		}
		w.Write([]byte(`{"data":[]}`))
	})
	items, err := geckoterminal.Collect(c.Pools.AllTopByNetwork(context.Background(), "eth", &geckoterminal.PoolListOptions{
		Sort: geckoterminal.Ptr(geckoterminal.SortH24VolumeUSDDesc),
	}))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(items) != 2 || calls.Load() != 2 {
		t.Fatalf("items=%d calls=%d", len(items), calls.Load())
	}
}

func TestClientProKeepsCustomRateLimit(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	// Explicit rate limit applied before WithCoinGeckoPro is kept.
	c, err := geckoterminal.NewClient(
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithCoinGeckoPro("key"),
		geckoterminal.WithBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
}
