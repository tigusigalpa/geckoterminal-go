// Package geckoterminal provides an idiomatic, zero-dependency Go client
// for the GeckoTerminal API v2 (CoinGecko Onchain DEX data).
//
// The client covers all 20 public endpoints: networks, DEXes, pools,
// tokens, trades, OHLCV, simple prices and pool search. It implements
// JSON:API 1.0 responses, exact decimal handling, client-side rate
// limiting, retries with exponential backoff, page and iterator based
// pagination, and a CoinGecko Pro authenticated mode.
//
// # Quick start
//
//	client, err := geckoterminal.NewClient()
//	if err != nil {
//		log.Fatal(err)
//	}
//	pools, err := client.Pools.Trending(ctx, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, p := range pools.Items {
//		fmt.Println(p.Attributes.Name)
//	}
//
// # Monetary values
//
// The API returns monetary values as JSON strings, sometimes with 60+
// digits after the decimal point. They are modeled as Decimal, which
// preserves the original string. Use Decimal.Float64 for display and
// Decimal.BigFloat for precise math.
//
// # Errors
//
// Non-2xx responses are returned as *APIError. Use errors.Is with
// ErrNotFound and ErrRateLimited, or the IsNotFound and IsRateLimited
// helpers. Client-side validation failures return *ValidationError.
//
// # CoinGecko Pro
//
// Use WithCoinGeckoPro to switch to the authenticated CoinGecko Pro
// on-chain API (https://pro-api.coingecko.com/api/v3/onchain) with a
// higher default rate limit.
//
// API documentation: https://api.geckoterminal.com/docs/index.html
// CoinGecko Onchain docs: https://docs.coingecko.com/reference/onchain-pools-overview
//
// License: MIT. Copyright (c) 2026 Igor Sazonov.
package geckoterminal
