package geckoterminal

import (
	"context"
	"iter"
	"net/url"
)

// TokensService covers the token endpoints.
type TokensService struct {
	client *Client
}

// Get returns a single token by address.
//
// GeckoTerminal: GET /networks/{network}/tokens/{address}
// Docs: https://docs.coingecko.com/reference/token-data-contract-address
func (s *TokensService) Get(ctx context.Context, network, address string, opts *TokenDetailOptions) (*Response[Token], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateAddress("address", address); err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts != nil {
		if err := setInclude(q, opts.Include, includeTokenTopPools); err != nil {
			return nil, err
		}
		setBool(q, "include_composition", opts.IncludeComposition)
		setBool(q, "include_inactive_source", opts.IncludeInactiveSource)
	}
	return fetchOne[Token](ctx, s.client, joinPath("networks", esc(network), "tokens", esc(address)), q)
}

// Multi returns multiple tokens by address (up to 30).
//
// GeckoTerminal: GET /networks/{network}/tokens/multi/{addresses}
// Docs: https://docs.coingecko.com/reference/tokens-data-contract-addresses
func (s *TokensService) Multi(ctx context.Context, network string, addresses []string, opts *TokenMultiOptions) (*Page[Token], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	joined, err := validateAddresses("addresses", addresses)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts != nil {
		if err := setInclude(q, opts.Include, includeTokenTopPools); err != nil {
			return nil, err
		}
		setBool(q, "include_composition", opts.IncludeComposition)
		setBool(q, "include_inactive_source", opts.IncludeInactiveSource)
	}
	p, err := fetchPage[Token](ctx, s.client, joinPath("networks", esc(network), "tokens", "multi", joined), q, 1)
	if err != nil {
		return nil, err
	}
	p.HasNext = false
	return p, nil
}

// Info returns extended metadata for a token: images, socials, GT
// score, holder distribution and honeypot flags.
//
// GeckoTerminal: GET /networks/{network}/tokens/{address}/info
// Docs: https://docs.coingecko.com/reference/token-info-contract-address
func (s *TokensService) Info(ctx context.Context, network, address string) (*Response[TokenInfo], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateAddress("address", address); err != nil {
		return nil, err
	}
	return fetchOne[TokenInfo](ctx, s.client, joinPath("networks", esc(network), "tokens", esc(address), "info"), nil)
}

// RecentlyUpdated returns up to 100 tokens whose metadata was updated
// most recently.
//
// GeckoTerminal: GET /tokens/info_recently_updated.
// Docs: https://docs.coingecko.com/reference/tokens-info-recent-updated
func (s *TokensService) RecentlyUpdated(ctx context.Context, opts *RecentlyUpdatedOptions) (*Page[TokenInfo], error) {
	q := url.Values{}
	if opts != nil {
		if opts.IncludeNetwork != nil && *opts.IncludeNetwork {
			q.Set("include", string(IncludeNetwork))
		}
		if opts.Network != nil && *opts.Network != "" {
			if err := validateNetwork(*opts.Network); err != nil {
				return nil, err
			}
			q.Set("network", *opts.Network)
		}
	}
	p, err := fetchPage[TokenInfo](ctx, s.client, "/tokens/info_recently_updated", q, 1)
	if err != nil {
		return nil, err
	}
	p.HasNext = false
	return p, nil
}

// AllRecentlyUpdated iterates over the recently updated tokens. The
// endpoint returns a single page, so the iterator yields at most one
// page of results.
//
// GeckoTerminal: GET /tokens/info_recently_updated.
func (s *TokensService) AllRecentlyUpdated(ctx context.Context, opts *RecentlyUpdatedOptions) iter.Seq2[TokenInfo, error] {
	return All(ctx, func(ctx context.Context, _ int) (*Page[TokenInfo], error) {
		return s.RecentlyUpdated(ctx, opts)
	})
}
