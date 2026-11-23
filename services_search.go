package geckoterminal

import (
	"context"
	"net/url"
)

// SearchService covers the pool search endpoint.
type SearchService struct {
	client *Client
}

// Pools searches pools by query, optionally filtered to a network. The
// query is required: although the OpenAPI schema marks it optional, a
// search without a query is meaningless.
//
// GeckoTerminal: GET /search/pools
// Docs: https://docs.coingecko.com/reference/search-pools
func (s *SearchService) Pools(ctx context.Context, query string, opts *SearchPoolsOptions) (*Page[Pool], error) {
	if err := validateSegment("query", query, maxSearchQuery); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("query", query)
	page := 1
	if opts != nil {
		if opts.Network != nil && *opts.Network != "" {
			if err := validateNetwork(*opts.Network); err != nil {
				return nil, err
			}
			q.Set("network", *opts.Network)
		}
		if err := setInclude(q, opts.Include, includePoolStandard); err != nil {
			return nil, err
		}
		if err := setPage(q, opts.Page); err != nil {
			return nil, err
		}
		page = Deref(opts.Page)
	}
	if page == 0 {
		page = 1
	}
	return fetchPage[Pool](ctx, s.client, "/search/pools", q, page)
}
