package geckoterminal

import (
	"context"
	"net/url"
	"strconv"
)

// TradesService covers the pool trades endpoint.
type TradesService struct {
	client *Client
}

// List returns the most recent trades of a pool (up to 300 over the
// last 24 hours on the public API). The trading_period, cursor and
// per_page parameters require CoinGecko Pro. When cursor pagination is
// used, Page.NextCursor returns the continuation token from
// meta.next_cursor.
//
// GeckoTerminal: GET /networks/{network}/pools/{pool_address}/trades
// Docs: https://docs.coingecko.com/reference/pool-trades-contract-address
func (s *TradesService) List(ctx context.Context, network, poolAddress string, opts *TradesOptions) (*Page[Trade], error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateAddress("pool_address", poolAddress); err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts != nil {
		if opts.TradeVolumeInUSDGreaterThan != nil {
			q.Set("trade_volume_in_usd_greater_than", opts.TradeVolumeInUSDGreaterThan.String())
		}
		setBool(q, "include_anomalous", opts.IncludeAnomalous)
		if err := setEnum(q, "trading_period", opts.TradingPeriod, validTradingPeriods); err != nil {
			return nil, err
		}
		setString(q, "cursor", opts.Cursor)
		if err := setLimit(q, "per_page", opts.PerPage, 1, maxTradesPerPage); err != nil {
			return nil, err
		}
		if err := setTokenParam(q, opts.Token); err != nil {
			return nil, err
		}
	}
	p, err := fetchPage[Trade](ctx, s.client, joinPath("networks", esc(network), "pools", esc(poolAddress), "trades"), q, 1)
	if err != nil {
		return nil, err
	}
	if p.NextCursor() == "" && p.Links == nil {
		p.HasNext = false
	}
	return p, nil
}

// int64Param applies an optional integer query parameter.
func int64Param(q url.Values, field string, v *int64) {
	if v != nil {
		q.Set(field, strconv.FormatInt(*v, 10))
	}
}
