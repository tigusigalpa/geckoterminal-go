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

func TestPoolsTrending(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_trending.json", func(r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
	}))
	page, err := c.Pools.Trending(context.Background(), nil)
	if err != nil {
		t.Fatalf("Trending: %v", err)
	}
	if gotPath != "/networks/trending_pools" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "" {
		t.Fatalf("query = %q", gotQuery)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
	p := page.Items[0]
	if p.Attributes.Name != "USDC / WETH 0.05%" {
		t.Fatalf("name = %q", p.Attributes.Name)
	}
	if p.Attributes.MarketCapUSD != nil {
		t.Fatal("MarketCapUSD should be nil")
	}
	if p.Attributes.VolumeUSD.H24() != "95000000.12" {
		t.Fatalf("volume h24 = %q", p.Attributes.VolumeUSD.H24())
	}
	if p.Attributes.Transactions.H24().Buys != 1200 {
		t.Fatalf("tx h24 buys = %d", p.Attributes.Transactions.H24().Buys)
	}
	sol := page.Items[1]
	if sol.Attributes.PoolCreatedAt != nil {
		t.Fatal("PoolCreatedAt should be nil")
	}
	if sol.Attributes.SentimentVotePositivePercentage == nil ||
		*sol.Attributes.SentimentVotePositivePercentage != 88.0 {
		t.Fatal("sentiment missing")
	}
	if sol.Attributes.CommunitySusReport == nil || *sol.Attributes.CommunitySusReport != 2 {
		t.Fatal("community_sus_report missing")
	}
	if len(page.Included.Tokens) != 1 || len(page.Included.Dexes) != 1 {
		t.Fatalf("included = %+v", page.Included)
	}
}

func TestPoolsTrendingAllOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "pools_trending.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.Pools.Trending(context.Background(), &geckoterminal.TrendingPoolsOptions{
		Page:                   geckoterminal.Ptr(2),
		Include:                []geckoterminal.Include{geckoterminal.IncludeBaseToken, geckoterminal.IncludeNetwork, geckoterminal.IncludeBaseToken},
		IncludeGTCommunityData: geckoterminal.Ptr(true),
		Duration:               geckoterminal.Ptr(geckoterminal.Duration1h),
	})
	if err != nil {
		t.Fatalf("Trending: %v", err)
	}
	if q.Get("page") != "2" {
		t.Fatalf("page = %q", q.Get("page"))
	}
	if q.Get("include") != "base_token,network" {
		t.Fatalf("include = %q", q.Get("include"))
	}
	if q.Get("include_gt_community_data") != "true" {
		t.Fatalf("gt = %q", q.Get("include_gt_community_data"))
	}
	if q.Get("duration") != "1h" {
		t.Fatalf("duration = %q", q.Get("duration"))
	}
}

func TestPoolsTrendingByNetwork(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_trending.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	page, err := c.Pools.TrendingByNetwork(context.Background(), "eth", nil)
	if err != nil {
		t.Fatalf("TrendingByNetwork: %v", err)
	}
	if gotPath != "/networks/eth/trending_pools" {
		t.Fatalf("path = %q", gotPath)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
}

func TestPoolsTrendingByNetworkIncludeNetworkRejected(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	_, err := c.Pools.TrendingByNetwork(context.Background(), "eth", &geckoterminal.TrendingPoolsOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeNetwork},
	})
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestPoolsGet(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "pool_detail.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	resp, err := c.Pools.Get(context.Background(), "eth", "0x60594a405d53811d3bc4766596efd80fd545a270", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/networks/eth/pools/0x60594a405d53811d3bc4766596efd80fd545a270" {
		t.Fatalf("path = %q", gotPath)
	}
	a := resp.Data.Attributes
	if a.Name != "DAI / WETH 0.05%" || a.PoolName != "DAI / WETH" {
		t.Fatalf("name = %q pool_name = %q", a.Name, a.PoolName)
	}
	if a.PoolFeePercentage != "0.05" {
		t.Fatalf("fee = %q", a.PoolFeePercentage)
	}
	if a.FDVUSD == nil || *a.FDVUSD != "4355394853.81743" {
		t.Fatalf("fdv = %v", a.FDVUSD)
	}
	if a.PoolCreatedAt == nil || a.PoolCreatedAt.Year() != 2021 {
		t.Fatalf("created_at = %v", a.PoolCreatedAt)
	}
	if a.LockedLiquidityPercentage != nil {
		t.Fatal("locked_liquidity_percentage should be nil")
	}
	if a.BaseTokenLiquidityUSD != "415145.065009079040209753105859666959720400776461442432" {
		t.Fatalf("base liquidity precision lost: %q", a.BaseTokenLiquidityUSD)
	}
	if a.NetBuyVolumeUSD.H24() != "-4716.5197581542" {
		t.Fatalf("net buy h24 = %q", a.NetBuyVolumeUSD.H24())
	}
	if a.BuyVolumeUSD.M5() != "573.6712893456" {
		t.Fatalf("buy m5 = %q", a.BuyVolumeUSD.M5())
	}
	if a.SellVolumeUSD.M5() != "0.0" {
		t.Fatalf("sell m5 = %q", a.SellVolumeUSD.M5())
	}
	if a.Transactions.M5().UniqueTraders != 2 {
		t.Fatalf("tx m5 = %+v", a.Transactions.M5())
	}
	if a.PriceChangePercentage.H6() != "-0.027" {
		t.Fatalf("price change h6 = %q", a.PriceChangePercentage.H6())
	}
	if a.SentimentVoteNegativePercentage == nil || *a.SentimentVoteNegativePercentage != 26.5 {
		t.Fatal("negative sentiment missing")
	}
}

func TestPoolsGetAllOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "pool_detail.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.Pools.Get(context.Background(), "eth", "0xabc", &geckoterminal.PoolDetailOptions{
		Include:                []geckoterminal.Include{geckoterminal.IncludeDEX},
		IncludeVolumeBreakdown: geckoterminal.Ptr(true),
		IncludeComposition:     geckoterminal.Ptr(false),
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if q.Get("include") != "dex" {
		t.Fatalf("include = %q", q.Get("include"))
	}
	if q.Get("include_volume_breakdown") != "true" {
		t.Fatalf("vol = %q", q.Get("include_volume_breakdown"))
	}
	if q.Get("include_composition") != "false" {
		t.Fatalf("comp = %q", q.Get("include_composition"))
	}
}

func TestPoolsMulti(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "pool_multi.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	page, err := c.Pools.Multi(context.Background(), "eth", []string{
		"0x60594a405d53811d3bc4766596efd80fd545a270",
		"0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
	}, nil)
	if err != nil {
		t.Fatalf("Multi: %v", err)
	}
	if gotPath != "/networks/eth/pools/multi/0x60594a405d53811d3bc4766596efd80fd545a270,0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640" {
		t.Fatalf("path = %q", gotPath)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.HasNext {
		t.Fatal("HasNext should be false for multi")
	}
}

func TestPoolsMultiAddressLimits(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{"data":[]}`))
	})
	// 30 addresses is allowed.
	addrs := make([]string, 30)
	for i := range addrs {
		addrs[i] = "0xaaa"
	}
	if _, err := c.Pools.Multi(context.Background(), "eth", addrs, nil); err != nil {
		t.Fatalf("30 addresses: %v", err)
	}
	// 31 is rejected client-side.
	_, err := c.Pools.Multi(context.Background(), "eth", append(addrs, "0xbbb"), nil)
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("31 addresses: err = %v", err)
	}
	// Empty list is rejected.
	if _, err := c.Pools.Multi(context.Background(), "eth", nil, nil); err == nil {
		t.Fatal("empty addresses: expected error")
	}
	// Empty element is rejected.
	if _, err := c.Pools.Multi(context.Background(), "eth", []string{"0xaaa", ""}, nil); err == nil {
		t.Fatal("empty element: expected error")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestPoolsTopByNetwork(t *testing.T) {
	t.Parallel()
	var gotPath, sort string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_trending.json", func(r *http.Request) {
		gotPath = r.URL.Path
		sort = r.URL.Query().Get("sort")
	}))
	_, err := c.Pools.TopByNetwork(context.Background(), "eth", &geckoterminal.PoolListOptions{
		Sort: geckoterminal.Ptr(geckoterminal.SortH24TxCountDesc),
	})
	if err != nil {
		t.Fatalf("TopByNetwork: %v", err)
	}
	if gotPath != "/networks/eth/pools" {
		t.Fatalf("path = %q", gotPath)
	}
	if sort != "h24_tx_count_desc" {
		t.Fatalf("sort = %q", sort)
	}
}

func TestPoolsTopByNetworkSortRejected(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	_, err := c.Pools.TopByNetwork(context.Background(), "eth", &geckoterminal.PoolListOptions{
		Sort: geckoterminal.Ptr(geckoterminal.SortH24VolumeUSDLiquidityDesc),
	})
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
}

func TestPoolsTopByDEX(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_trending.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	_, err := c.Pools.TopByDEX(context.Background(), "eth", "uniswap_v3", &geckoterminal.PoolListOptions{
		Page:    geckoterminal.Ptr(1),
		Include: []geckoterminal.Include{geckoterminal.IncludeQuoteToken},
	})
	if err != nil {
		t.Fatalf("TopByDEX: %v", err)
	}
	if gotPath != "/networks/eth/dexes/uniswap_v3/pools" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestPoolsTopByToken(t *testing.T) {
	t.Parallel()
	var gotPath, sort, inactive string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_trending.json", func(r *http.Request) {
		gotPath = r.URL.Path
		sort = r.URL.Query().Get("sort")
		inactive = r.URL.Query().Get("include_inactive_source")
	}))
	_, err := c.Pools.TopByToken(context.Background(), "eth", "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", &geckoterminal.TokenPoolsOptions{
		Sort:                  geckoterminal.Ptr(geckoterminal.SortH24VolumeUSDLiquidityDesc),
		IncludeInactiveSource: geckoterminal.Ptr(true),
	})
	if err != nil {
		t.Fatalf("TopByToken: %v", err)
	}
	if gotPath != "/networks/eth/tokens/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48/pools" {
		t.Fatalf("path = %q", gotPath)
	}
	if sort != "h24_volume_usd_liquidity_desc" {
		t.Fatalf("sort = %q", sort)
	}
	if inactive != "true" {
		t.Fatalf("inactive = %q", inactive)
	}
}

func TestPoolsNew(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_new.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	page, err := c.Pools.New(context.Background(), &geckoterminal.NewPoolsOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeNetwork},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if gotPath != "/networks/new_pools" {
		t.Fatalf("path = %q", gotPath)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
}

func TestPoolsNewByNetwork(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_new.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	page, err := c.Pools.NewByNetwork(context.Background(), "base", nil)
	if err != nil {
		t.Fatalf("NewByNetwork: %v", err)
	}
	if gotPath != "/networks/base/new_pools" {
		t.Fatalf("path = %q", gotPath)
	}
	if page.Items[0].Attributes.Name != "MEME / WETH" {
		t.Fatalf("name = %q", page.Items[0].Attributes.Name)
	}
}

func TestPoolsNewByNetworkIncludeNetworkRejected(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	_, err := c.Pools.NewByNetwork(context.Background(), "eth", &geckoterminal.NewPoolsOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeNetwork},
	})
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
}

func TestPoolsTokensInfo(t *testing.T) {
	t.Parallel()
	var gotPath, include string
	c, _ := newTestClient(t, fixtureHandler(t, "pool_tokens_info.json", func(r *http.Request) {
		gotPath = r.URL.Path
		include = r.URL.Query().Get("include")
	}))
	page, err := c.Pools.TokensInfo(context.Background(), "eth", "0x60594a405d53811d3bc4766596efd80fd545a270",
		&geckoterminal.PoolTokensInfoOptions{IncludePool: geckoterminal.Ptr(true)})
	if err != nil {
		t.Fatalf("TokensInfo: %v", err)
	}
	if gotPath != "/networks/eth/pools/0x60594a405d53811d3bc4766596efd80fd545a270/info" {
		t.Fatalf("path = %q", gotPath)
	}
	if include != "pool" {
		t.Fatalf("include = %q", include)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[0].Attributes.Symbol != "DAI" {
		t.Fatalf("symbol = %q", page.Items[0].Attributes.Symbol)
	}
	if len(page.Included.Pools) != 1 {
		t.Fatalf("included pools = %d", len(page.Included.Pools))
	}
}

func TestPoolsTokensInfoNoInclude(t *testing.T) {
	t.Parallel()
	var rawQuery string
	c, _ := newTestClient(t, fixtureHandler(t, "pool_tokens_info.json", func(r *http.Request) {
		rawQuery = r.URL.RawQuery
	}))
	if _, err := c.Pools.TokensInfo(context.Background(), "eth", "0xabc", nil); err != nil {
		t.Fatalf("TokensInfo: %v", err)
	}
	if rawQuery != "" {
		t.Fatalf("query = %q", rawQuery)
	}
}

func TestPoolsValidationErrors(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	ctx := context.Background()
	check := func(name string, err error) {
		t.Helper()
		var ve *geckoterminal.ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("%s: err = %v, want *ValidationError", name, err)
		}
	}
	_, e1 := c.Pools.TopByNetwork(ctx, "", nil)
	check("empty network", e1)
	_, e2 := c.Pools.Get(ctx, "eth", "bad/addr", nil)
	check("bad address", e2)
	_, e3 := c.Pools.Trending(ctx, &geckoterminal.TrendingPoolsOptions{
		Duration: geckoterminal.Ptr(geckoterminal.Duration("2h")),
	})
	check("duration", e3)
	_, e4 := c.Pools.TopByNetwork(ctx, "eth", &geckoterminal.PoolListOptions{
		Include: []geckoterminal.Include{geckoterminal.Include("bogus")},
	})
	check("bad include", e4)
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestPools404(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, statusHandler(http.StatusNotFound, string(fixture(t, "error_404.json")), nil))
	_, err := c.Pools.Get(context.Background(), "eth", "0xdeadbeef", nil)
	if !geckoterminal.IsNotFound(err) {
		t.Fatalf("err = %v", err)
	}
}

func TestPoolsAllTopByNetwork(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		page := r.URL.Query().Get("page")
		if page == "1" || page == "" {
			w.Write(fixture(t, "pools_trending.json"))
			return
		}
		w.Write([]byte(`{"data":[]}`))
	})
	items, err := geckoterminal.Collect(c.Pools.AllTopByNetwork(context.Background(), "eth", nil))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	// Page 1 has 2 items; page 2 is empty and stops iteration.
	if len(items) != 2 {
		t.Fatalf("items = %d", len(items))
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestPoolsPathEscaping(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "pools_trending.json", func(r *http.Request) {
		gotPath = r.URL.RawQuery
	}))
	_, err := c.Pools.TopByNetwork(context.Background(), "eth", &geckoterminal.PoolListOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeBaseToken, geckoterminal.IncludeQuoteToken},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(gotPath, "include=base_token%2Cquote_token") && !strings.Contains(gotPath, "include=base_token,quote_token") {
		t.Fatalf("query = %q", gotPath)
	}
}
