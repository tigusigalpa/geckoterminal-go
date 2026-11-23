# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-11-23

### Added

- `Client` with eight services (`Networks`, `Dexes`, `Pools`, `Tokens`,
  `Trades`, `OHLCV`, `Simple`, `Search`) covering all 20 public
  GeckoTerminal API v2 endpoints.
- Functional options: `WithHTTPClient`, `WithBaseURL`, `WithAPIVersion`,
  `WithUserAgent`, `WithCoinGeckoPro`, `WithAPIKey`, `WithRateLimit`,
  `WithRetryPolicy`, `WithTimeout`, `WithLogger`, `WithClock`,
  `WithCache`.
- JSON:API 1.0 envelope: `Response[T]`, `Page[T]`, `Links`, `Included`
  with typed lookup helpers, `Relationship` with `ID()`/`IDs()`.
- `Decimal` type preserving the exact string form of monetary values,
  with `Float64`, `MustFloat64`, `BigFloat`, `IsZero` and full JSON
  round-trip support (strings, numbers and null).
- Typed models for pools, tokens, token info, trades, OHLCV candles
  (`Candle` decodes the `[ts, o, h, l, c, v]` array form), networks,
  DEXes and the simple price maps.
- Client-side token-bucket rate limiter (10 req/min public, 250 req/min
  Pro) with lazy refill, injected clock support and context-aware waits.
- Retry with exponential backoff, jitter, `Retry-After` support
  (seconds and HTTP-date) and a configurable `RetryPolicy`.
- Cursor- and links-aware `Page.HasNext`, `Page.NextCursor()` for
  CoinGecko Pro trades, plus `All`/`Collect` iterators built on
  `iter.Seq2`.
- CoinGecko Pro mode (`x-cg-pro-api-key` header, `/onchain` path
  prefix via the Pro base URL).
- Optional HTTP response cache: `Cache` interface and `MemoryCache`
  LRU implementation with `Cache-Control` TTL and ETag revalidation.
- Client-side validation with `*ValidationError` for every endpoint:
  path segments, page bounds, multi-address limits (30), OHLCV
  timeframe/aggregate matrix, `per_page`, enum and include sets.
- `*APIError` with `errors.Is`/`errors.As` support, `IsNotFound`,
  `IsRateLimited`, `ErrNotFound`, `ErrRateLimited`,
  `ErrInvalidArgument`.
- `log/slog` diagnostics (debug request logs, warn on retries) with a
  no-op default logger; credentials and response bodies are never
  logged.
- Golden `testdata/` fixtures for every endpoint, unit tests on
  `httptest`, fuzz targets for `Decimal`, `Candle` and `Included`,
  benchmarks, and build-tagged integration tests.
- Examples under `examples/` for networks, trending pools, OHLCV,
  simple prices and pagination.
- GitHub Actions CI (test matrix, lint, govulncheck), release and
  CodeQL workflows.

### Notes

- The `include_composition` parameter is accepted and forwarded, but
  the response shape is undocumented upstream and therefore not modeled.
- `launchpad_details` is kept as `map[string]any` because its shape is
  undocumented.
- The public API is in beta (`v2-beta`); the SDK pins
  `Accept: application/json;version=20230203` by default, configurable
  via `WithAPIVersion`.

[Unreleased]: https://github.com/tigusigalpa/geckoterminal-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/tigusigalpa/geckoterminal-go/releases/tag/v0.1.0
