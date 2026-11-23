package geckoterminal

import (
	"context"
	"iter"
	"net/url"
)

// PoolsService covers the pool endpoints.
type PoolsService struct {
	client *Client
}

// Trending returns the trending pools across all networks.
//
// GeckoTerminal: GET /networks/trending_pools.
// Docs: https://docs.coingecko.com/reference/trending-pools-list
func (s *PoolsService) Trending(ctx context.Context, opts *TrendingPoolsOptions) (*Page[Pool], error) {
	q, page, err := trendingQuery(opts, includePoolWithNetwork)
	if err != nil {
		return nil, err
	}
	return fetchPage[Pool](ctx, s.client, "/networks/trending_pools", q, page)
}

// TrendingByNetwork returns the trending pools of a single network.
//
// GeckoTerminal: GET /networks/{network}/trending_pools
// Docs: https://docs.coingecko.com/reference/trending-pools-network
func (s *PoolsService) TrendingByNetwork(ctx context.Context, network string, opts *TrendingPoolsOptions) (*Page[Pool], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	q, page, err := trendingQuery(opts, includePoolStandard)
	if err != nil {
		return nil, err
	}
	return fetchPage[Pool](ctx, s.client, joinPath("networks", esc(network), "trending_pools"), q, page)
}

// AllTrending iterates over all pages of the trending pools.
//
// GeckoTerminal: GET /networks/trending_pools.
func (s *PoolsService) AllTrending(ctx context.Context, opts *TrendingPoolsOptions) iter.Seq2[Pool, error] {
	return All(ctx, func(ctx context.Context, page int) (*Page[Pool], error) {
		o := &TrendingPoolsOptions{Page: Ptr(page)}
		if opts != nil {
			*o = *opts
			o.Page = Ptr(page)
		}
		return s.Trending(ctx, o)
	})
}

// trendingQuery validates trending options and builds the query. The
// all-networks variant additionally accepts "network" in include.
func trendingQuery(opts *TrendingPoolsOptions, allowedInclude map[Include]bool) (url.Values, int, error) {
	q := url.Values{}
	page := 1
	if opts == nil {
		return q, page, nil
	}
	if err := setPage(q, opts.Page); err != nil {
		return nil, 0, err
	}
	if err := setInclude(q, opts.Include, allowedInclude); err != nil {
		return nil, 0, err
	}
	setBool(q, "include_gt_community_data", opts.IncludeGTCommunityData)
	if err := setEnum(q, "duration", opts.Duration, validDurations); err != nil {
		return nil, 0, err
	}
	if opts.Page != nil {
		page = *opts.Page
	}
	return q, page, nil
}

// Get returns a single pool by address.
//
// GeckoTerminal: GET /networks/{network}/pools/{address}
// Docs: https://docs.coingecko.com/reference/pool-address
func (s *PoolsService) Get(ctx context.Context, network, poolAddress string, opts *PoolDetailOptions) (*Response[Pool], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateAddress("address", poolAddress); err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts != nil {
		if err := setInclude(q, opts.Include, includePoolStandard); err != nil {
			return nil, err
		}
		setBool(q, "include_volume_breakdown", opts.IncludeVolumeBreakdown)
		setBool(q, "include_composition", opts.IncludeComposition)
	}
	return fetchOne[Pool](ctx, s.client, joinPath("networks", esc(network), "pools", esc(poolAddress)), q)
}

// Multi returns multiple pools by address (up to 30).
//
// GeckoTerminal: GET /networks/{network}/pools/multi/{addresses}
// Docs: https://docs.coingecko.com/reference/pools-addresses
func (s *PoolsService) Multi(ctx context.Context, network string, addresses []string, opts *MultiPoolOptions) (*Page[Pool], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	joined, err := validateAddresses("addresses", addresses)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts != nil {
		if err := setInclude(q, opts.Include, includePoolStandard); err != nil {
			return nil, err
		}
		setBool(q, "include_volume_breakdown", opts.IncludeVolumeBreakdown)
		setBool(q, "include_composition", opts.IncludeComposition)
	}
	p, err := fetchPage[Pool](ctx, s.client, joinPath("networks", esc(network), "pools", "multi", joined), q, 1)
	if err != nil {
		return nil, err
	}
	p.HasNext = false
	return p, nil
}

// TopByNetwork returns the top pools of a network.
//
// GeckoTerminal: GET /networks/{network}/pools.
// Docs: https://docs.coingecko.com/reference/top-pools-network
func (s *PoolsService) TopByNetwork(ctx context.Context, network string, opts *PoolListOptions) (*Page[Pool], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	q, page, err := poolListQuery(opts)
	if err != nil {
		return nil, err
	}
	return fetchPage[Pool](ctx, s.client, joinPath("networks", esc(network), "pools"), q, page)
}

// AllTopByNetwork iterates over all pages of the top pools of a
// network.
//
// GeckoTerminal: GET /networks/{network}/pools.
func (s *PoolsService) AllTopByNetwork(ctx context.Context, network string, opts *PoolListOptions) iter.Seq2[Pool, error] {
	return All(ctx, func(ctx context.Context, page int) (*Page[Pool], error) {
		o := &PoolListOptions{Page: Ptr(page)}
		if opts != nil {
			*o = *opts
			o.Page = Ptr(page)
		}
		return s.TopByNetwork(ctx, network, o)
	})
}

// poolListQuery validates shared pool list options.
func poolListQuery(opts *PoolListOptions) (url.Values, int, error) {
	q := url.Values{}
	page := 1
	if opts == nil {
		return q, page, nil
	}
	if err := setPage(q, opts.Page); err != nil {
		return nil, 0, err
	}
	if err := setInclude(q, opts.Include, includePoolStandard); err != nil {
		return nil, 0, err
	}
	setBool(q, "include_gt_community_data", opts.IncludeGTCommunityData)
	if err := setEnum(q, "sort", opts.Sort, sortPoolStandard); err != nil {
		return nil, 0, err
	}
	if opts.Page != nil {
		page = *opts.Page
	}
	return q, page, nil
}

// TopByDEX returns the top pools of a DEX on a network.
//
// GeckoTerminal: GET /networks/{network}/dexes/{dex}/pools
// Docs: https://docs.coingecko.com/reference/top-pools-dex
func (s *PoolsService) TopByDEX(ctx context.Context, network, dex string, opts *PoolListOptions) (*Page[Pool], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateSegment("dex", dex, 128); err != nil {
		return nil, err
	}
	q, page, err := poolListQuery(opts)
	if err != nil {
		return nil, err
	}
	return fetchPage[Pool](ctx, s.client, joinPath("networks", esc(network), "dexes", esc(dex), "pools"), q, page)
}

// TopByToken returns the top pools containing a token.
//
// GeckoTerminal: GET /networks/{network}/tokens/{token_address}/pools
// Docs: https://docs.coingecko.com/reference/top-pools-contract-address
func (s *PoolsService) TopByToken(ctx context.Context, network, tokenAddress string, opts *TokenPoolsOptions) (*Page[Pool], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateAddress("token_address", tokenAddress); err != nil {
		return nil, err
	}
	q := url.Values{}
	page := 1
	if opts != nil {
		if err := setPage(q, opts.Page); err != nil {
			return nil, err
		}
		if err := setInclude(q, opts.Include, includePoolStandard); err != nil {
			return nil, err
		}
		setBool(q, "include_inactive_source", opts.IncludeInactiveSource)
		if err := setEnum(q, "sort", opts.Sort, sortByToken); err != nil {
			return nil, err
		}
		page = Deref(opts.Page)
	}
	if page == 0 {
		page = 1
	}
	return fetchPage[Pool](ctx, s.client, joinPath("networks", esc(network), "tokens", esc(tokenAddress), "pools"), q, page)
}

// New returns the newest pools across all networks.
//
// GeckoTerminal: GET /networks/new_pools
// Docs: https://docs.coingecko.com/reference/latest-pools-list
func (s *PoolsService) New(ctx context.Context, opts *NewPoolsOptions) (*Page[Pool], error) {
	q, page, err := newPoolsQuery(opts, includePoolWithNetwork)
	if err != nil {
		return nil, err
	}
	return fetchPage[Pool](ctx, s.client, "/networks/new_pools", q, page)
}

// NewByNetwork returns the newest pools of a network.
//
// GeckoTerminal: GET /networks/{network}/new_pools
// Docs: https://docs.coingecko.com/reference/latest-pools-network
func (s *PoolsService) NewByNetwork(ctx context.Context, network string, opts *NewPoolsOptions) (*Page[Pool], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	q, page, err := newPoolsQuery(opts, includePoolStandard)
	if err != nil {
		return nil, err
	}
	return fetchPage[Pool](ctx, s.client, joinPath("networks", esc(network), "new_pools"), q, page)
}

// newPoolsQuery validates shared new-pools options.
func newPoolsQuery(opts *NewPoolsOptions, allowedInclude map[Include]bool) (url.Values, int, error) {
	q := url.Values{}
	page := 1
	if opts == nil {
		return q, page, nil
	}
	if err := setPage(q, opts.Page); err != nil {
		return nil, 0, err
	}
	if err := setInclude(q, opts.Include, allowedInclude); err != nil {
		return nil, 0, err
	}
	setBool(q, "include_gt_community_data", opts.IncludeGTCommunityData)
	if opts.Page != nil {
		page = *opts.Page
	}
	return q, page, nil
}

// TokensInfo returns extended metadata for the base and quote tokens
// of a pool.
//
// GeckoTerminal: GET /networks/{network}/pools/{pool_address}/info
// Docs: https://docs.coingecko.com/reference/pool-token-info-contract-address
func (s *PoolsService) TokensInfo(ctx context.Context, network, poolAddress string, opts *PoolTokensInfoOptions) (*Page[TokenInfo], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateAddress("pool_address", poolAddress); err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts != nil && opts.IncludePool != nil && *opts.IncludePool {
		q.Set("include", string(IncludePool))
	}
	return fetchPage[TokenInfo](ctx, s.client, joinPath("networks", esc(network), "pools", esc(poolAddress), "info"), q, 1)
}
