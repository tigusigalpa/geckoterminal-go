# geckoterminal-go

[![Go Reference](https://pkg.go.dev/badge/github.com/tigusigalpa/geckoterminal-go.svg)](https://pkg.go.dev/github.com/tigusigalpa/geckoterminal-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/tigusigalpa/geckoterminal-go)](https://goreportcard.com/report/github.com/tigusigalpa/geckoterminal-go)
[![CI](https://github.com/tigusigalpa/geckoterminal-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tigusigalpa/geckoterminal-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23%2B-blue)](https://go.dev/dl/)

**Tigusigalpa/GeckoTerminal** — idiomatic, zero-dependency Go SDK for the
[GeckoTerminal](https://www.geckoterminal.com/dex-api) (CoinGecko Onchain)
API v2. GeckoTerminal provides on-chain DEX data: pools, tokens, trades
and OHLCV across 200+ networks and 1500+ DEXes.

## Features

- All **20** public endpoints of the GeckoTerminal API v2, fully typed
- JSON:API 1.0 envelope with typed `Included` resources and relationship
  lookup helpers
- `Decimal` type that preserves monetary values verbatim — no `float64`
  precision loss on strings like
  `716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726`
- Client-side token-bucket rate limiting (10 req/min public, 250 req/min
  CoinGecko Pro)
- Retries with exponential backoff, jitter and `Retry-After` support
- Page-based and iterator-based pagination (`iter.Seq2`)
- CoinGecko Pro mode (`x-cg-pro-api-key`, `/onchain` path prefix)
- Optional in-memory response cache with `Cache-Control` TTL and ETag
  revalidation
- Typed errors: `*APIError` with `errors.Is`/`errors.As`, sentinel
  errors, `*ValidationError`
- `context` support everywhere, `log/slog` diagnostics, race-safe
- **Zero runtime dependencies** — standard library only, Go 1.23+

## Installation

```sh
go get github.com/tigusigalpa/geckoterminal-go
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := geckoterminal.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	pools, err := client.Pools.Trending(ctx, &geckoterminal.TrendingPoolsOptions{
		Duration: geckoterminal.Ptr(geckoterminal.Duration24h),
		Include:  []geckoterminal.Include{geckoterminal.IncludeBaseToken, geckoterminal.IncludeDEX},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range pools.Items {
		fmt.Printf("%-40s %s USD\n", p.Attributes.Name, p.Attributes.BaseTokenPriceUSD)
	}
}
```

## Authentication & rate limits

| Mode | Base URL | Credentials | Limit |
| --- | --- | --- | --- |
| Public (default) | `https://api.geckoterminal.com/api/v2` | none | ~10 req/min |
| CoinGecko Pro | `https://pro-api.coingecko.com/api/v3/onchain` | `x-cg-pro-api-key` header | 250 req/min |

The public API is in **beta** and can change without notice. The SDK
pins `Accept: application/json;version=20230203` by default; override
with `WithAPIVersion`. All responses are cached server-side for ~1
minute, and on-chain data is typically available 2–3 seconds after a
transaction is confirmed.

## Endpoint coverage

| # | SDK method | HTTP | Docs |
| --- | --- | --- | --- |
| 1 | `client.Networks.List` | `GET /networks` | [link](https://docs.coingecko.com/reference/networks-list) |
| 2 | `client.Dexes.ListByNetwork` | `GET /networks/{network}/dexes` | [link](https://docs.coingecko.com/reference/dexes-list) |
| 3 | `client.Pools.Trending` | `GET /networks/trending_pools` | [link](https://docs.coingecko.com/reference/trending-pools-list) |
| 4 | `client.Pools.TrendingByNetwork` | `GET /networks/{network}/trending_pools` | [link](https://docs.coingecko.com/reference/trending-pools-network) |
| 5 | `client.Pools.Get` | `GET /networks/{network}/pools/{address}` | [link](https://docs.coingecko.com/reference/pool-address) |
| 6 | `client.Pools.Multi` | `GET /networks/{network}/pools/multi/{addresses}` | [link](https://docs.coingecko.com/reference/pools-addresses) |
| 7 | `client.Pools.TopByNetwork` | `GET /networks/{network}/pools` | [link](https://docs.coingecko.com/reference/top-pools-network) |
| 8 | `client.Pools.TopByDEX` | `GET /networks/{network}/dexes/{dex}/pools` | [link](https://docs.coingecko.com/reference/top-pools-dex) |
| 9 | `client.Pools.TopByToken` | `GET /networks/{network}/tokens/{token_address}/pools` | [link](https://docs.coingecko.com/reference/top-pools-contract-address) |
| 10 | `client.Pools.New` | `GET /networks/new_pools` | [link](https://docs.coingecko.com/reference/latest-pools-list) |
| 11 | `client.Pools.NewByNetwork` | `GET /networks/{network}/new_pools` | [link](https://docs.coingecko.com/reference/latest-pools-network) |
| 12 | `client.Pools.TokensInfo` | `GET /networks/{network}/pools/{pool_address}/info` | [link](https://docs.coingecko.com/reference/pool-token-info-contract-address) |
| 13 | `client.Tokens.Get` | `GET /networks/{network}/tokens/{address}` | [link](https://docs.coingecko.com/reference/token-data-contract-address) |
| 14 | `client.Tokens.Multi` | `GET /networks/{network}/tokens/multi/{addresses}` | [link](https://docs.coingecko.com/reference/tokens-data-contract-addresses) |
| 15 | `client.Tokens.Info` | `GET /networks/{network}/tokens/{address}/info` | [link](https://docs.coingecko.com/reference/token-info-contract-address) |
| 16 | `client.Tokens.RecentlyUpdated` | `GET /tokens/info_recently_updated` | [link](https://docs.coingecko.com/reference/tokens-info-recent-updated) |
| 17 | `client.Trades.List` | `GET /networks/{network}/pools/{pool_address}/trades` | [link](https://docs.coingecko.com/reference/pool-trades-contract-address) |
| 18 | `client.OHLCV.Get` | `GET /networks/{network}/pools/{pool_address}/ohlcv/{timeframe}` | [link](https://docs.coingecko.com/reference/pool-ohlcv-contract-address) |
| 19 | `client.Simple.TokenPrices` | `GET /simple/networks/{network}/token_price/{addresses}` | [link](https://docs.coingecko.com/reference/onchain-simple-price) |
| 20 | `client.Search.Pools` | `GET /search/pools` | [link](https://docs.coingecko.com/reference/search-pools) |

## Usage examples

### Networks

```go
page, err := client.Networks.List(ctx, &geckoterminal.ListOptions{
	Page: geckoterminal.Ptr(1),
})
if err != nil {
	log.Fatal(err)
}
for _, n := range page.Items {
	fmt.Println(n.ID, n.Attributes.Name)
}
```

### DEXes

```go
dexes, err := client.Dexes.ListByNetwork(ctx, "eth", nil)
```

### Pools

```go
// Single pool with related resources.
pool, err := client.Pools.Get(ctx, "eth",
	"0x60594a405d53811d3bc4766596efd80fd545a270",
	&geckoterminal.PoolDetailOptions{
		Include:                []geckoterminal.Include{geckoterminal.IncludeBaseToken, geckoterminal.IncludeQuoteToken},
		IncludeVolumeBreakdown: geckoterminal.Ptr(true),
	})
if err != nil {
	log.Fatal(err)
}
if tok, ok := pool.Included.Token(pool.Data.Relationships.BaseToken.ID()); ok {
	fmt.Println("base token:", tok.Attributes.Symbol)
}

// Up to 30 pools in one request.
pools, err := client.Pools.Multi(ctx, "eth", []string{addr1, addr2}, nil)
```

### Tokens

```go
token, err := client.Tokens.Get(ctx, "eth",
	"0xdac17f958d2ee523a2206206994597c13d831ec7",
	&geckoterminal.TokenDetailOptions{
		Include: []geckoterminal.Include{geckoterminal.IncludeTopPools},
	})

info, err := client.Tokens.Info(ctx, "eth",
	"0xdac17f958d2ee523a2206206994597c13d831ec7")
if err == nil && info.Data.Attributes.Holders != nil {
	fmt.Println("holders:", info.Data.Attributes.Holders.Count)
}
```

### Trades

```go
trades, err := client.Trades.List(ctx, "eth", poolAddress,
	&geckoterminal.TradesOptions{
		TradeVolumeInUSDGreaterThan: geckoterminal.Ptr(geckoterminal.Decimal("100")),
	})
```

### OHLCV

```go
resp, err := client.OHLCV.Get(ctx, "eth", poolAddress,
	geckoterminal.TimeframeHour, &geckoterminal.OHLCVOptions{
		Aggregate: geckoterminal.Ptr(geckoterminal.Aggregate4),
		Limit:     geckoterminal.Ptr(100),
	})
if err != nil {
	log.Fatal(err)
}
for _, c := range resp.Data.Attributes.OHLCVList {
	fmt.Println(c.Timestamp, c.Close)
}

// One request covers at most 6 months of history. To go deeper, pass
// the oldest timestamp of the previous response minus one:
opts := &geckoterminal.OHLCVOptions{
	BeforeTimestamp: geckoterminal.Ptr(oldest - 1),
}
```

### Simple prices

```go
prices, err := client.Simple.TokenPrices(ctx, "eth",
	[]string{"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"},
	&geckoterminal.TokenPricesOptions{
		IncludeMarketCap: geckoterminal.Ptr(true),
	})
if err != nil {
	log.Fatal(err)
}
fmt.Println(prices.Data.TokenPrices["0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"])
```

### Search

```go
results, err := client.Search.Pools(ctx, "weth",
	&geckoterminal.SearchPoolsOptions{
		Network: geckoterminal.Ptr("eth"),
	})
```

## Pagination & iterators

List endpoints return `*Page[T]` with `Items`, `Number`, `HasNext` and
optional JSON:API `Links`. Page numbers run 1..10.

```go
for pool, err := range client.Pools.AllTrending(ctx, nil) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pool.Attributes.Name)
}

// Drain an iterator into a slice:
networks, err := geckoterminal.Collect(client.Networks.All(ctx, nil))
```

Iterator methods available: `Networks.All`, `Dexes.AllByNetwork`,
`Pools.AllTrending`, `Pools.AllTopByNetwork`,
`Tokens.AllRecentlyUpdated`. The generic `geckoterminal.All` helper can
wrap any `fetch(ctx, page)` function.

CoinGecko Pro trades use cursor pagination: pass `TradesOptions.Cursor`
and continue with `page.NextCursor()`.

## Decimal precision

The API returns monetary values as strings to avoid floating-point
loss. The SDK models them as `Decimal`, which keeps the original string
verbatim and accepts JSON strings, numbers and `null`.

```go
d := pool.Attributes.ReserveInUSD // Decimal

fmt.Println(d.String())          // exact original value
f, _ := d.Float64()              // for display / loose math
bf, ok := d.BigFloat()           // arbitrary precision
if !d.IsZero() {
	fmt.Println(bf.Text('f', 10))
}
```

Nullable numeric fields (`fdv_usd`, `market_cap_usd`,
`locked_liquidity_percentage`) are `*Decimal`; nullable strings and
timestamps use `*string` and `*time.Time`.

## Error handling

```go
pools, err := client.Pools.TopByNetwork(ctx, "eth", nil)
switch {
case err == nil:
	// ok
case geckoterminal.IsNotFound(err):
	// HTTP 404
case geckoterminal.IsRateLimited(err):
	// HTTP 429
	default:
		var apiErr *geckoterminal.APIError
		if errors.As(err, &apiErr) {
			fmt.Println(apiErr.StatusCode, apiErr.RefID, apiErr.RequestID, apiErr.RetryAfter)
		}
		var valErr *geckoterminal.ValidationError
		if errors.As(err, &valErr) {
			fmt.Println(valErr.Field, valErr.Value, valErr.Message)
		}
}
```

- `*APIError` — non-2xx responses. `Unwrap` maps 404 → `ErrNotFound`
  and 429 → `ErrRateLimited`. The raw body (up to 64 KiB) is kept in
  `Body` for diagnostics and may be non-JSON.
- `*ValidationError` — client-side failures raised **before** any HTTP
  request. Unwraps to `ErrInvalidArgument`.

## Retries, timeouts, rate limiting

```go
client, err := geckoterminal.NewClient(
	geckoterminal.WithTimeout(15*time.Second),
	geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Jitter:     0.3,
		RetryOn:    []int{429, 500, 502, 503, 504},
	}),
	geckoterminal.WithRateLimit(30, 5), // 30 req/min, burst 5
)
```

Retries apply to GET requests and the `RetryOn` status codes only;
`Retry-After` takes precedence over backoff; `context` cancellation is
respected while sleeping. The rate limiter is a shared token bucket —
a `429` from the server does not consume extra allowance.

## CoinGecko Pro mode

```go
client, err := geckoterminal.NewClient(
	geckoterminal.WithCoinGeckoPro(os.Getenv("COINGECKO_API_KEY")),
)
```

This sets the base URL to `https://pro-api.coingecko.com/api/v3/onchain`,
sends the `x-cg-pro-api-key` header and raises the default rate limit to
250 req/min. Apply `WithBaseURL` afterwards to keep the key but target a
custom gateway, or use `WithAPIKey` to set only the header.

The API key is never placed in URLs, logs or error bodies.

## Caching

```go
client, err := geckoterminal.NewClient(
	geckoterminal.WithCache(geckoterminal.NewMemoryCache(256)),
)
```

`MemoryCache` is a bounded LRU that honors `Cache-Control` (`s-maxage`,
`max-age`, `no-store`, `no-cache`, `private`) and revalidates expired
entries with `If-None-Match`/`ETag`. Implement the `Cache` interface for
your own backend. Errors are never cached.

## Versioning

The SDK follows [SemVer](https://semver.org). Until `v1.0.0`, minor
releases may introduce breaking changes. The `Accept` header pins the
API version `20230203` by default (`WithAPIVersion` to change it),
isolating you from upstream API version bumps.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Bug reports and feature
requests: [GitHub Issues](https://github.com/tigusigalpa/geckoterminal-go/issues).

## License

[MIT](LICENSE). Copyright (c) 2026 Igor Sazonov.
