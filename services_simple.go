package geckoterminal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SimpleService covers the simple token price endpoint.
type SimpleService struct {
	client *Client
}

// TokenPrices returns USD prices for up to 30 token addresses. The
// response Data holds the flattened attribute maps keyed by token
// address.
//
// GeckoTerminal: GET /simple/networks/{network}/token_price/{addresses}
// Docs: https://docs.coingecko.com/reference/onchain-simple-price
func (s *SimpleService) TokenPrices(ctx context.Context, network string, addresses []string,
	opts *TokenPricesOptions) (*Response[TokenPrices], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	joined, err := validateAddresses("addresses", addresses)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts != nil {
		setBool(q, "include_market_cap", opts.IncludeMarketCap)
		setBool(q, "mcap_fdv_fallback", opts.McapFDVFallback)
		setBool(q, "include_24hr_vol", opts.Include24hVol)
		setBool(q, "include_24hr_price_change", opts.Include24hPriceChange)
		setBool(q, "include_total_reserve_in_usd", opts.IncludeTotalReserveInUSD)
		setBool(q, "include_inactive_source", opts.IncludeInactiveSource)
	}
	path := joinPath("simple", "networks", esc(network), "token_price", joined)
	body, _, err := s.client.do(ctx, http.MethodGet, path, q)
	if err != nil {
		return nil, err
	}
	var env struct {
		Data struct {
			ID         string      `json:"id"`
			Type       string      `json:"type"`
			Attributes TokenPrices `json:"attributes"`
		} `json:"data"`
		Included Included        `json:"included"`
		Links    *Links          `json:"links"`
		Meta     json.RawMessage `json:"meta"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("geckoterminal: decode %s: %w", path, err)
	}
	return &Response[TokenPrices]{
		Data:     env.Data.Attributes,
		Included: env.Included,
		Links:    env.Links,
		Meta:     env.Meta,
	}, nil
}
