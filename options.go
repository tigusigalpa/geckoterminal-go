package geckoterminal

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// Duration is the trending pools lookback window.
type Duration string

// Trending pool duration windows.
const (
	Duration5m  Duration = "5m"
	Duration1h  Duration = "1h"
	Duration6h  Duration = "6h"
	Duration24h Duration = "24h"
)

// String returns the API representation of the duration.
func (d Duration) String() string { return string(d) }

// Valid reports whether the duration is a valid trending window.
func (d Duration) Valid() bool { return validDurations[d] }

var validDurations = map[Duration]bool{
	Duration5m: true, Duration1h: true, Duration6h: true, Duration24h: true,
}

// Sort is a pool list ordering.
type Sort string

// Pool list orderings. Note that h24_volume_usd_liquidity_desc is only
// accepted by the "top pools by token" endpoint.
const (
	SortH24VolumeUSDDesc          Sort = "h24_volume_usd_desc"
	SortH24TxCountDesc            Sort = "h24_tx_count_desc"
	SortH24VolumeUSDLiquidityDesc Sort = "h24_volume_usd_liquidity_desc"
)

// String returns the API representation of the sort ordering.
func (s Sort) String() string { return string(s) }

var (
	sortPoolStandard = map[Sort]bool{
		SortH24VolumeUSDDesc: true, SortH24TxCountDesc: true,
	}
	sortByToken = map[Sort]bool{
		SortH24VolumeUSDDesc: true, SortH24TxCountDesc: true, SortH24VolumeUSDLiquidityDesc: true,
	}
)

// Include is a JSON:API related resource requested via the "include"
// query parameter. The set of accepted values is endpoint specific.
type Include string

// Supported include values.
const (
	IncludeBaseToken  Include = "base_token"
	IncludeQuoteToken Include = "quote_token"
	IncludeDEX        Include = "dex"
	IncludeNetwork    Include = "network"
	IncludeTopPools   Include = "top_pools"
	IncludePool       Include = "pool"
)

// String returns the API representation of the include value.
func (i Include) String() string { return string(i) }

var (
	includePoolStandard = map[Include]bool{
		IncludeBaseToken: true, IncludeQuoteToken: true, IncludeDEX: true,
	}
	includePoolWithNetwork = map[Include]bool{
		IncludeBaseToken: true, IncludeQuoteToken: true, IncludeDEX: true, IncludeNetwork: true,
	}
	includeTokenTopPools = map[Include]bool{IncludeTopPools: true}
)

// TradingPeriod is the trades lookback window (CoinGecko Pro only).
type TradingPeriod string

// Trades lookback windows.
const (
	TradingPeriod1d  TradingPeriod = "1d"
	TradingPeriod7d  TradingPeriod = "7d"
	TradingPeriod30d TradingPeriod = "30d"
)

// String returns the API representation of the trading period.
func (t TradingPeriod) String() string { return string(t) }

var validTradingPeriods = map[TradingPeriod]bool{
	TradingPeriod1d: true, TradingPeriod7d: true, TradingPeriod30d: true,
}

// Ptr returns a pointer to v.
func Ptr[T any](v T) *T { return &v }

// Deref returns the value pointed to by p, or the zero value if p is nil.
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// config holds the resolved client configuration.
type config struct {
	httpClient   *http.Client
	baseURL      *url.URL
	apiVersion   string
	userAgent    string
	apiKey       string
	rateLimit    RateLimit
	rateLimitSet bool
	retry        RetryPolicy
	timeout      time.Duration
	logger       *slog.Logger
	now          func() time.Time
	cache        Cache
}

func defaultConfig() config {
	u, _ := url.Parse(DefaultBaseURL)
	return config{
		baseURL:    u,
		apiVersion: DefaultAPIVersion,
		userAgent:  DefaultUserAgent,
		rateLimit:  DefaultRateLimit,
		retry:      DefaultRetryPolicy,
		timeout:    30 * time.Second,
	}
}

// Option configures a Client. Options are applied in order; the last one
// wins for overlapping settings.
type Option func(*config) error

// WithHTTPClient replaces the HTTP client entirely. Use it to customize
// transports, proxies or TLS. When set, WithTimeout has no effect: the
// caller controls timeouts on the provided client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) error {
		if hc == nil {
			return &ValidationError{Field: "httpClient", Value: "nil", Message: "must not be nil"}
		}
		c.httpClient = hc
		return nil
	}
}

// WithBaseURL overrides the API base URL. Useful for tests, proxies and
// self-hosted gateways. When used after WithCoinGeckoPro it takes
// precedence over the Pro base URL while keeping the API key header.
func WithBaseURL(rawurl string) Option {
	return func(c *config) error {
		if rawurl == "" {
			return &ValidationError{Field: "baseURL", Value: "", Message: "must not be empty"}
		}
		u, err := url.Parse(rawurl)
		if err != nil {
			return fmt.Errorf("geckoterminal: invalid base URL: %w", err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return &ValidationError{Field: "baseURL", Value: rawurl, Message: "scheme must be http or https"}
		}
		if u.Host == "" {
			return &ValidationError{Field: "baseURL", Value: rawurl, Message: "missing host"}
		}
		u.RawQuery = ""
		u.Fragment = ""
		c.baseURL = u
		return nil
	}
}

// WithAPIVersion sets the "version=" parameter of the Accept header.
// An empty string disables the version parameter entirely.
func WithAPIVersion(v string) Option {
	return func(c *config) error {
		c.apiVersion = v
		return nil
	}
}

// WithUserAgent overrides the User-Agent header. An empty string keeps
// the default.
func WithUserAgent(ua string) Option {
	return func(c *config) error {
		if ua != "" {
			c.userAgent = ua
		}
		return nil
	}
}

// WithCoinGeckoPro switches the client to the authenticated CoinGecko
// Pro on-chain API: base URL https://pro-api.coingecko.com/api/v3/onchain,
// the x-cg-pro-api-key header, and a 250 req/min default rate limit
// (unless WithRateLimit was applied earlier). Apply WithBaseURL after
// WithCoinGeckoPro to keep the key but target a custom Pro gateway.
func WithCoinGeckoPro(apiKey string) Option {
	return func(c *config) error {
		if apiKey == "" {
			return &ValidationError{Field: "apiKey", Value: "", Message: "must not be empty"}
		}
		u, err := url.Parse(ProBaseURL)
		if err != nil {
			return err
		}
		c.baseURL = u
		c.apiKey = apiKey
		if !c.rateLimitSet {
			c.rateLimit = DefaultProRateLimit
		}
		return nil
	}
}

// WithAPIKey sets the x-cg-pro-api-key header without changing the base
// URL. Useful with custom Pro proxies. The public GeckoTerminal API does
// not accept an API key.
func WithAPIKey(apiKey string) Option {
	return func(c *config) error {
		if apiKey == "" {
			return &ValidationError{Field: "apiKey", Value: "", Message: "must not be empty"}
		}
		c.apiKey = apiKey
		return nil
	}
}

// WithRateLimit sets a custom client-side rate limit. A
// requestsPerMinute <= 0 disables the limiter.
func WithRateLimit(requestsPerMinute float64, burst int) Option {
	return func(c *config) error {
		c.rateLimit = RateLimit{RequestsPerMinute: requestsPerMinute, Burst: burst}
		c.rateLimitSet = true
		return nil
	}
}

// WithRetryPolicy sets a custom retry policy. Unset fields are
// normalized to DefaultRetryPolicy values; set MaxRetries to 0 to
// disable retries.
func WithRetryPolicy(p RetryPolicy) Option {
	return func(c *config) error {
		c.retry = p.normalize()
		return nil
	}
}

// WithTimeout sets the HTTP request timeout on the internally created
// http.Client. It has no effect when WithHTTPClient is used.
func WithTimeout(d time.Duration) Option {
	return func(c *config) error {
		if d <= 0 {
			return &ValidationError{Field: "timeout", Value: d.String(), Message: "must be positive"}
		}
		c.timeout = d
		return nil
	}
}

// WithLogger sets a slog.Logger for request diagnostics. nil keeps the
// default no-op logger. The SDK never logs request credentials or
// response bodies.
func WithLogger(l *slog.Logger) Option {
	return func(c *config) error {
		c.logger = l
		return nil
	}
}

// WithClock injects the clock used by the rate limiter. Intended for
// deterministic tests; production code should leave the default.
func WithClock(now func() time.Time) Option {
	return func(c *config) error {
		if now == nil {
			return &ValidationError{Field: "clock", Value: "nil", Message: "must not be nil"}
		}
		c.now = now
		return nil
	}
}

// WithCache installs an HTTP response cache. GET responses carrying
// cache-friendly Cache-Control headers are stored and reused; entries
// with an ETag are revalidated via If-None-Match. nil disables caching.
func WithCache(cache Cache) Option {
	return func(c *config) error {
		c.cache = cache
		return nil
	}
}

// Request option structs. All fields are pointers (or nil-able slices)
// so that "unset" is distinguishable from the zero value; nil opts
// equals all-default options.

// ListOptions paginates the networks list.
type ListOptions struct {
	// Page selects the result page, 1..10.
	Page *int
}

// DEXListOptions paginates the DEX list of a network.
type DEXListOptions struct {
	// Page selects the result page, 1..10.
	Page *int
}

// PoolListOptions configures pool list endpoints.
type PoolListOptions struct {
	// Page selects the result page, 1..10.
	Page *int
	// Include requests related resources (base_token, quote_token, dex).
	Include []Include
	// IncludeGTCommunityData adds GeckoTerminal community metrics.
	IncludeGTCommunityData *bool
	// Sort orders results (h24_volume_usd_desc or h24_tx_count_desc).
	Sort *Sort
}

// TokenPoolsOptions configures the "top pools by token" endpoint, which
// additionally supports include_inactive_source and a third sort mode.
type TokenPoolsOptions struct {
	// Page selects the result page, 1..10.
	Page *int
	// Include requests related resources (base_token, quote_token, dex).
	Include []Include
	// IncludeInactiveSource includes pools from inactive sources.
	IncludeInactiveSource *bool
	// Sort orders results, including h24_volume_usd_liquidity_desc.
	Sort *Sort
}

// TrendingPoolsOptions configures the trending pools endpoints.
type TrendingPoolsOptions struct {
	// Page selects the result page, 1..10.
	Page *int
	// Include requests related resources (base_token, quote_token, dex;
	// the all-networks variant also accepts network).
	Include []Include
	// IncludeGTCommunityData adds GeckoTerminal community metrics.
	IncludeGTCommunityData *bool
	// Duration selects the trending window (5m, 1h, 6h, 24h).
	Duration *Duration
}

// PoolDetailOptions configures the single-pool endpoint.
type PoolDetailOptions struct {
	// Include requests related resources (base_token, quote_token, dex).
	Include []Include
	// IncludeVolumeBreakdown adds buy/sell/net volume metrics.
	IncludeVolumeBreakdown *bool
	// IncludeComposition requests the pool composition object.
	// Its response shape is undocumented.
	IncludeComposition *bool
}

// MultiPoolOptions configures the multi-pool endpoint.
type MultiPoolOptions struct {
	// Include requests related resources (base_token, quote_token, dex).
	Include []Include
	// IncludeVolumeBreakdown adds buy/sell/net volume metrics.
	IncludeVolumeBreakdown *bool
	// IncludeComposition requests the pool composition object.
	IncludeComposition *bool
}

// NewPoolsOptions configures the new-pools endpoints.
type NewPoolsOptions struct {
	// Page selects the result page, 1..10.
	Page *int
	// Include requests related resources (base_token, quote_token, dex;
	// the all-networks variant also accepts network).
	Include []Include
	// IncludeGTCommunityData adds GeckoTerminal community metrics.
	IncludeGTCommunityData *bool
}

// SearchPoolsOptions configures pool search.
type SearchPoolsOptions struct {
	// Network filters results to a single network.
	Network *string
	// Include requests related resources (base_token, quote_token, dex).
	Include []Include
	// Page selects the result page, 1..10.
	Page *int
}

// PoolTokensInfoOptions configures the pool tokens info endpoint.
type PoolTokensInfoOptions struct {
	// IncludePool includes the pool resource in "included".
	IncludePool *bool
}

// TokenDetailOptions configures the single-token endpoint.
type TokenDetailOptions struct {
	// Include requests related resources (top_pools).
	Include []Include
	// IncludeComposition requests the token composition object.
	IncludeComposition *bool
	// IncludeInactiveSource includes pools from inactive sources.
	IncludeInactiveSource *bool
}

// TokenMultiOptions configures the multi-token endpoint.
type TokenMultiOptions struct {
	// Include requests related resources (top_pools).
	Include []Include
	// IncludeComposition requests the token composition object.
	IncludeComposition *bool
	// IncludeInactiveSource includes pools from inactive sources.
	IncludeInactiveSource *bool
}

// RecentlyUpdatedOptions configures the recently updated tokens list.
type RecentlyUpdatedOptions struct {
	// IncludeNetwork includes each token's network in "included".
	IncludeNetwork *bool
	// Network filters results to a single network.
	Network *string
}

// TradesOptions configures the pool trades endpoint.
type TradesOptions struct {
	// TradeVolumeInUSDGreaterThan filters trades below the USD volume.
	TradeVolumeInUSDGreaterThan *Decimal
	// IncludeAnomalous includes anomalous trades.
	IncludeAnomalous *bool
	// TradingPeriod selects the lookback window (CoinGecko Pro only).
	TradingPeriod *TradingPeriod
	// Cursor continues a cursor-paginated listing (CoinGecko Pro only).
	Cursor *string
	// PerPage sets the page size, 1..300 (CoinGecko Pro only).
	PerPage *int
	// Token inverts trade data: "base", "quote" or a token address.
	Token *string
}

// OHLCVOptions configures the pool OHLCV endpoint.
type OHLCVOptions struct {
	// Aggregate sets the aggregation factor. Valid values depend on the
	// timeframe: day -> 1; hour -> 1, 4, 12; minute -> 1, 5, 15.
	Aggregate *Aggregate
	// BeforeTimestamp returns candles before this Unix timestamp.
	// Required to page deeper than 6 months of history.
	BeforeTimestamp *int64
	// Limit is the number of candles, 1..1000 (default 100).
	Limit *int
	// Currency selects the price currency (usd or token).
	Currency *Currency
	// IncludeEmptyIntervals returns intervals without trades.
	IncludeEmptyIntervals *bool
	// Token inverts the chart: "base", "quote" or a token address.
	Token *string
}

// TokenPricesOptions configures the simple token price endpoint.
type TokenPricesOptions struct {
	// IncludeMarketCap adds per-token market capitalization.
	IncludeMarketCap *bool
	// McapFDVFallback falls back to FDV when market cap is missing.
	McapFDVFallback *bool
	// Include24hVol adds per-token 24h volume.
	Include24hVol *bool
	// Include24hPriceChange adds per-token 24h price change percentage.
	Include24hPriceChange *bool
	// IncludeTotalReserveInUSD adds per-token total reserve.
	IncludeTotalReserveInUSD *bool
	// IncludeInactiveSource includes pools from inactive sources.
	IncludeInactiveSource *bool
}
