package geckoterminal

import (
	"context"
	"iter"
	"net/url"
)

// DexesService covers the supported DEX endpoints.
type DexesService struct {
	client *Client
}

// ListByNetwork returns the supported DEXes on a network.
//
// GeckoTerminal: GET /networks/{network}/dexes.
// Docs: https://docs.coingecko.com/reference/dexes-list
func (s *DexesService) ListByNetwork(ctx context.Context, network string, opts *DEXListOptions) (*Page[DEX], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	q := url.Values{}
	page := 1
	if opts != nil {
		if err := setPage(q, opts.Page); err != nil {
			return nil, err
		}
		page = Deref(opts.Page)
	}
	if page == 0 {
		page = 1
	}
	return fetchPage[DEX](ctx, s.client, joinPath("networks", esc(network), "dexes"), q, page)
}

// AllByNetwork iterates over all pages of the DEX list of a network.
//
// GeckoTerminal: GET /networks/{network}/dexes.
func (s *DexesService) AllByNetwork(ctx context.Context, network string, opts *DEXListOptions) iter.Seq2[DEX, error] {
	return All(ctx, func(ctx context.Context, page int) (*Page[DEX], error) {
		o := &DEXListOptions{Page: Ptr(page)}
		if opts != nil {
			*o = *opts
			o.Page = Ptr(page)
		}
		return s.ListByNetwork(ctx, network, o)
	})
}
