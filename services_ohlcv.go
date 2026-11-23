package geckoterminal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OHLCVService covers the pool candle endpoint.
type OHLCVService struct {
	client *Client
}

// Get returns OHLCV candles for a pool. A single request covers at
// most 6 months of history; pass OHLCVOptions.BeforeTimestamp (the
// oldest timestamp of the previous response minus one) to page deeper.
//
// GeckoTerminal: GET /networks/{network}/pools/{pool_address}/ohlcv/{timeframe}.
// Docs: https://docs.coingecko.com/reference/pool-ohlcv-contract-address
func (s *OHLCVService) Get(ctx context.Context, network, poolAddress string, timeframe Timeframe,
	opts *OHLCVOptions) (*OHLCVResponse, error) {
	if err := validateNetwork(network); err != nil {
		return nil, err
	}
	if err := validateAddress("pool_address", poolAddress); err != nil {
		return nil, err
	}
	if !timeframe.Valid() {
		return nil, &ValidationError{
			Field:   "timeframe",
			Value:   string(timeframe),
			Message: "must be one of: day, hour, minute",
		}
	}
	q := url.Values{}
	if opts != nil {
		if opts.Aggregate != nil {
			if !opts.Aggregate.validFor(timeframe) {
				return nil, &ValidationError{
					Field:   "aggregate",
					Value:   string(*opts.Aggregate),
					Message: fmt.Sprintf("not valid for timeframe %q", timeframe),
				}
			}
			q.Set("aggregate", string(*opts.Aggregate))
		}
		if opts.BeforeTimestamp != nil {
			if *opts.BeforeTimestamp <= 0 {
				return nil, &ValidationError{
					Field:   "before_timestamp",
					Value:   fmt.Sprint(*opts.BeforeTimestamp),
					Message: "must be a positive Unix timestamp",
				}
			}
			int64Param(q, "before_timestamp", opts.BeforeTimestamp)
		}
		if err := setLimit(q, "limit", opts.Limit, 1, maxOHLCVLimit); err != nil {
			return nil, err
		}
		if err := setEnum(q, "currency", opts.Currency, validCurrencies); err != nil {
			return nil, err
		}
		setBool(q, "include_empty_intervals", opts.IncludeEmptyIntervals)
		if err := setTokenParam(q, opts.Token); err != nil {
			return nil, err
		}
	}
	path := joinPath("networks", esc(network), "pools", esc(poolAddress), "ohlcv", string(timeframe))
	body, _, err := s.client.do(ctx, http.MethodGet, path, q)
	if err != nil {
		return nil, err
	}
	var resp OHLCVResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("geckoterminal: decode %s: %w", path, err)
	}
	return &resp, nil
}

// Candles is a convenience wrapper around Get that returns just the
// candle list.
//
// GeckoTerminal: GET /networks/{network}/pools/{pool_address}/ohlcv/{timeframe}.
func (s *OHLCVService) Candles(ctx context.Context, network, poolAddress string, timeframe Timeframe,
	opts *OHLCVOptions) ([]Candle, error) {
	resp, err := s.Get(ctx, network, poolAddress, timeframe, opts)
	if err != nil {
		return nil, err
	}
	return resp.Data.Attributes.OHLCVList, nil
}
