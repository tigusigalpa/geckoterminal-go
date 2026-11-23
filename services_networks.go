package geckoterminal

import (
	"context"
	"iter"
	"net/url"
)

// NetworksService covers the supported networks endpoint.
type NetworksService struct {
	client *Client
}

// List returns the supported blockchain networks. It is the only
// endpoint that emits JSON:API pagination links.
//
// GeckoTerminal: GET /networks.
// Docs: https://docs.coingecko.com/reference/networks-list
func (s *NetworksService) List(ctx context.Context, opts *ListOptions) (*Page[Network], error) {
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
	return fetchPage[Network](ctx, s.client, "/networks", q, page)
}

// All iterates over all pages of the networks list.
//
// GeckoTerminal: GET /networks.
func (s *NetworksService) All(ctx context.Context, opts *ListOptions) iter.Seq2[Network, error] {
	return All(ctx, func(ctx context.Context, page int) (*Page[Network], error) {
		o := &ListOptions{Page: Ptr(page)}
		if opts != nil {
			*o = *opts
			o.Page = Ptr(page)
		}
		return s.List(ctx, o)
	})
}
