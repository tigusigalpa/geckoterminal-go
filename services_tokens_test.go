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

func TestTokensGet(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "token_detail.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	resp, err := c.Tokens.Get(context.Background(), "eth", "0xdac17f958d2ee523a2206206994597c13d831ec7", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotPath != "/networks/eth/tokens/0xdac17f958d2ee523a2206206994597c13d831ec7" {
		t.Fatalf("path = %q", gotPath)
	}
	a := resp.Data.Attributes
	if a.Symbol != "USDT" || a.Decimals != 6 {
		t.Fatalf("token = %+v", a)
	}
	if a.CoinGeckoCoinID == nil || *a.CoinGeckoCoinID != "tether" {
		t.Fatalf("coingecko id = %v", a.CoinGeckoCoinID)
	}
	if a.PriceUSD != "0.999789" {
		t.Fatalf("price = %q", a.PriceUSD)
	}
	if a.TotalReserveInUSD != "335714495.059134513199726364304069580849343577823934518315621723260820486809151975191568775052401150520637996530317619622915" {
		t.Fatalf("reserve precision lost: %q", a.TotalReserveInUSD)
	}
	if resp.Data.Relationships == nil || resp.Data.Relationships.TopPools == nil {
		t.Fatal("top_pools relationship missing")
	}
	if ids := resp.Data.Relationships.TopPools.IDs(); len(ids) != 2 {
		t.Fatalf("top pool ids = %v", ids)
	}
	if len(resp.Included.Pools) != 1 {
		t.Fatalf("included pools = %d", len(resp.Included.Pools))
	}
	if resp.Included.Pools[0].Attributes.TokenPriceUSD != "0.999789" {
		t.Fatalf("token_price_usd = %q", resp.Included.Pools[0].Attributes.TokenPriceUSD)
	}
}

func TestTokensGetOptions(t *testing.T) {
	t.Parallel()
	var q url.Values
	c, _ := newTestClient(t, fixtureHandler(t, "token_detail.json", func(r *http.Request) {
		q = r.URL.Query()
	}))
	_, err := c.Tokens.Get(context.Background(), "eth", "0xabc", &geckoterminal.TokenDetailOptions{
		Include:               []geckoterminal.Include{geckoterminal.IncludeTopPools},
		IncludeComposition:    geckoterminal.Ptr(true),
		IncludeInactiveSource: geckoterminal.Ptr(false),
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if q.Get("include") != "top_pools" {
		t.Fatalf("include = %q", q.Get("include"))
	}
	if q.Get("include_composition") != "true" || q.Get("include_inactive_source") != "false" {
		t.Fatalf("q = %v", q)
	}
}

func TestTokensGetIncludeRejected(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {})
	_, err := c.Tokens.Get(context.Background(), "eth", "0xabc", &geckoterminal.TokenDetailOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeDEX},
	})
	var ve *geckoterminal.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v", err)
	}
}

func TestTokensMulti(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "tokens_multi.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	page, err := c.Tokens.Multi(context.Background(), "eth", []string{
		"0xdac17f958d2ee523a2206206994597c13d831ec7",
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
	}, nil)
	if err != nil {
		t.Fatalf("Multi: %v", err)
	}
	if gotPath != "/networks/eth/tokens/multi/0xdac17f958d2ee523a2206206994597c13d831ec7,0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" {
		t.Fatalf("path = %q", gotPath)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[1].Attributes.Symbol != "WETH" {
		t.Fatalf("symbol = %q", page.Items[1].Attributes.Symbol)
	}
}

func TestTokensMultiValidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	addrs := make([]string, 31)
	for i := range addrs {
		addrs[i] = "0xaaa"
	}
	if _, err := c.Tokens.Multi(context.Background(), "eth", addrs, nil); err == nil {
		t.Fatal("31 addresses: expected error")
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestTokensInfo(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestClient(t, fixtureHandler(t, "token_info.json", func(r *http.Request) {
		gotPath = r.URL.Path
	}))
	resp, err := c.Tokens.Info(context.Background(), "eth", "0xdac17f958d2ee523a2206206994597c13d831ec7")
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if gotPath != "/networks/eth/tokens/0xdac17f958d2ee523a2206206994597c13d831ec7/info" {
		t.Fatalf("path = %q", gotPath)
	}
	a := resp.Data.Attributes
	if a.Symbol != "USDT" {
		t.Fatalf("symbol = %q", a.Symbol)
	}
	if a.Image == nil || a.Image.Large == "" {
		t.Fatal("image missing")
	}
	if a.GTScore != 92.6605504587156 {
		t.Fatalf("gt_score = %v", a.GTScore)
	}
	if a.GTScoreDetails == nil || a.GTScoreDetails.Pool != 87.5 {
		t.Fatalf("gt_score_details = %+v", a.GTScoreDetails)
	}
	if len(a.GTCategoriesID) != 2 {
		t.Fatalf("gt_categories_id = %v", a.GTCategoriesID)
	}
	if a.Holders == nil || a.Holders.Count != 7041203 {
		t.Fatalf("holders = %+v", a.Holders)
	}
	if a.Holders.DistributionPercentage.Top10 != "45.5782" {
		t.Fatalf("top10 = %q", a.Holders.DistributionPercentage.Top10)
	}
	if a.Holders.DistributionPercentage.P11To30 != "13.4293" {
		t.Fatalf("11_30 = %q", a.Holders.DistributionPercentage.P11To30)
	}
	if a.Holders.LastUpdated == nil || a.Holders.LastUpdated.Year() != 2025 {
		t.Fatalf("last_updated = %v", a.Holders.LastUpdated)
	}
	if a.IsHoneypot == nil || *a.IsHoneypot {
		t.Fatalf("is_honeypot = %v", a.IsHoneypot)
	}
	if a.DiscordURL != nil || a.TelegramHandle != nil {
		t.Fatal("nullable socials should be nil")
	}
	if a.TwitterHandle == nil || *a.TwitterHandle != "Tether_to" {
		t.Fatalf("twitter = %v", a.TwitterHandle)
	}
}

func TestTokensRecentlyUpdated(t *testing.T) {
	t.Parallel()
	var gotPath, network, include string
	c, _ := newTestClient(t, fixtureHandler(t, "tokens_recently_updated.json", func(r *http.Request) {
		gotPath = r.URL.Path
		network = r.URL.Query().Get("network")
		include = r.URL.Query().Get("include")
	}))
	page, err := c.Tokens.RecentlyUpdated(context.Background(), &geckoterminal.RecentlyUpdatedOptions{
		Network:        geckoterminal.Ptr("solana"),
		IncludeNetwork: geckoterminal.Ptr(true),
	})
	if err != nil {
		t.Fatalf("RecentlyUpdated: %v", err)
	}
	if gotPath != "/tokens/info_recently_updated" {
		t.Fatalf("path = %q", gotPath)
	}
	if network != "solana" || include != "network" {
		t.Fatalf("network=%q include=%q", network, include)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
	a := page.Items[0].Attributes
	if a.Symbol != "WOD" || a.Decimals != 6 {
		t.Fatalf("token = %+v", a)
	}
	if a.MetadataUpdatedAt == nil {
		t.Fatal("metadata_updated_at missing")
	}
	if len(a.GTCategoryIDs) != 1 || a.GTCategoryIDs[0] != "cat-9" {
		t.Fatalf("gt_category_ids = %v", a.GTCategoryIDs)
	}
	if page.Items[0].Relationships == nil || page.Items[0].Relationships.Network == nil ||
		page.Items[0].Relationships.Network.ID() != "solana" {
		t.Fatal("network relationship missing")
	}
	if len(page.Included.Networks) != 2 {
		t.Fatalf("included networks = %d", len(page.Included.Networks))
	}
	if page.HasNext {
		t.Fatal("HasNext should be false")
	}
}

func TestTokensRecentlyUpdatedAll(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, fixtureHandler(t, "tokens_recently_updated.json", nil))
	items, err := geckoterminal.Collect(c.Tokens.AllRecentlyUpdated(context.Background(), nil))
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d", len(items))
	}
}

func TestTokensValidation(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) })
	ctx := context.Background()
	if _, err := c.Tokens.Get(ctx, "", "0xabc", nil); err == nil {
		t.Fatal("empty network: expected error")
	}
	if _, err := c.Tokens.Get(ctx, "eth", "", nil); err == nil {
		t.Fatal("empty address: expected error")
	}
	if _, err := c.Tokens.Info(ctx, "eth", "a/b"); err == nil {
		t.Fatal("bad address: expected error")
	}
	bad := "bogus network"
	if _, err := c.Tokens.RecentlyUpdated(ctx, &geckoterminal.RecentlyUpdatedOptions{
		Network: &bad,
	}); err == nil {
		t.Fatal("bad filter network: expected error")
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d", calls.Load())
	}
}
