package geckoterminal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// maxBodyBytes caps the size of a successful response body. The
	// largest API responses (1000 candles, 300 trades) stay well below
	// this limit.
	maxBodyBytes = 32 << 20 // 32 MiB
	// maxErrorBodyBytes caps the diagnostic body stored in APIError.
	maxErrorBodyBytes = 64 << 10 // 64 KiB
	// defaultCacheTTL applies when a response lacks explicit
	// Cache-Control freshness information.
	defaultCacheTTL = 30 * time.Second
)

// Client is a GeckoTerminal API v2 client. It is safe for concurrent
// use. Construct it with NewClient and access endpoints through the
// exported service fields.
type Client struct {
	http       *http.Client
	baseURL    *url.URL
	apiVersion string
	userAgent  string
	apiKey     string
	limiter    *rateLimiter
	retry      RetryPolicy
	logger     *slog.Logger
	now        func() time.Time
	cache      Cache

	// Networks covers the supported networks list.
	Networks *NetworksService
	// Dexes covers the supported DEX list per network.
	Dexes *DexesService
	// Pools covers pool list, detail, multi and token-info endpoints.
	Pools *PoolsService
	// Tokens covers token detail, multi, info and recent-update endpoints.
	Tokens *TokensService
	// Trades covers the pool trades endpoint.
	Trades *TradesService
	// OHLCV covers the pool candle endpoint.
	OHLCV *OHLCVService
	// Simple covers the simple token price endpoint.
	Simple *SimpleService
	// Search covers pool search.
	Search *SearchService
}

// NewClient creates a client with sensible defaults:
//
//   - base URL https://api.geckoterminal.com/api/v2
//   - API version 20230203 in the Accept header
//   - client-side rate limit of 10 req/min, burst 10
//   - 3 retries with exponential backoff and jitter
//   - 30 second HTTP timeout
//
// Options are applied in order; the last one wins for overlapping
// settings. NewClient validates the configuration and returns an error
// instead of panicking.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	hc := cfg.httpClient
	if hc == nil {
		hc = &http.Client{Timeout: cfg.timeout}
	}
	logger := cfg.logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	now := cfg.now
	if now == nil {
		now = time.Now
	}
	c := &Client{
		http:       hc,
		baseURL:    cfg.baseURL,
		apiVersion: cfg.apiVersion,
		userAgent:  cfg.userAgent,
		apiKey:     cfg.apiKey,
		limiter:    newRateLimiter(cfg.rateLimit, now),
		retry:      cfg.retry.normalize(),
		logger:     logger,
		now:        now,
		cache:      cfg.cache,
	}
	c.Networks = &NetworksService{client: c}
	c.Dexes = &DexesService{client: c}
	c.Pools = &PoolsService{client: c}
	c.Tokens = &TokensService{client: c}
	c.Trades = &TradesService{client: c}
	c.OHLCV = &OHLCVService{client: c}
	c.Simple = &SimpleService{client: c}
	c.Search = &SearchService{client: c}
	return c, nil
}

// MustNewClient is like NewClient but panics on error. It is intended
// for tests and examples.
func MustNewClient(opts ...Option) *Client {
	c, err := NewClient(opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// acceptHeader builds the Accept header value for the configured API
// version.
func (c *Client) acceptHeader() string {
	if c.apiVersion == "" {
		return "application/json"
	}
	return "application/json;version=" + c.apiVersion
}

// buildURL combines the configured base URL with an endpoint path (whose
// parameters are already escaped) and the encoded query string.
func (c *Client) buildURL(endpointPath string, q url.Values) string {
	u := *c.baseURL
	escaped := strings.TrimSuffix(u.EscapedPath(), "/") + "/" + strings.TrimPrefix(endpointPath, "/")
	if decoded, err := url.PathUnescape(escaped); err == nil {
		u.Path = decoded
	}
	u.RawPath = escaped
	u.RawQuery = q.Encode()
	return u.String()
}

// joinPath joins URL path fragments into an absolute endpoint path.
// Elements must already be escaped where required.
func joinPath(elem ...string) string {
	return "/" + strings.Join(elem, "/")
}

// esc escapes a single path parameter.
func esc(v string) string { return url.PathEscape(v) }

// do performs an HTTP request and returns the response body and headers.
// It applies the rate limiter, retries per the configured policy, the
// optional response cache, and maps non-2xx responses to *APIError.
func (c *Client) do(ctx context.Context, method, endpointPath string, q url.Values) ([]byte, http.Header, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	reqURL := c.buildURL(endpointPath, q)

	var key string
	var cached CacheEntry
	var haveCached bool
	if c.cache != nil && method == http.MethodGet {
		key = cacheKey(c.apiVersion, reqURL)
		if entry, ok := c.cache.Get(key); ok {
			if entry.fresh(c.now()) {
				return entry.Body, entry.Header, nil
			}
			cached, haveCached = entry, true
		}
	}

	maxAttempts := 1 + c.retry.MaxRetries
	for attempt := 1; ; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, nil, err
		}
		req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("geckoterminal: build request: %w", err)
		}
		req.Header.Set("Accept", c.acceptHeader())
		req.Header.Set("User-Agent", c.userAgent)
		if c.apiKey != "" {
			req.Header.Set("x-cg-pro-api-key", c.apiKey)
		}
		if haveCached && cached.ETag != "" {
			req.Header.Set("If-None-Match", cached.ETag)
		}

		start := c.now()
		resp, err := c.http.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, nil, ctx.Err()
			}
			if attempt < maxAttempts {
				c.logRetry(endpointPath, 0, attempt, c.retry.delay(attempt, 0), err)
				if serr := sleepContext(ctx, c.retry.delay(attempt, 0)); serr != nil {
					return nil, nil, serr
				}
				continue
			}
			return nil, nil, fmt.Errorf("geckoterminal: %s %s: %w", method, endpointPath, err)
		}

		if resp.StatusCode == http.StatusNotModified && haveCached {
			_ = resp.Body.Close()
			if ttl := cacheTTL(resp.Header, defaultCacheTTL); ttl > 0 {
				c.cache.Set(key, cached, ttl)
			}
			return cached.Body, cached.Header, nil
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			body, err := readLimited(resp.Body, maxBodyBytes)
			_ = resp.Body.Close()
			if err != nil {
				return nil, nil, fmt.Errorf("geckoterminal: %s %s: %w", method, endpointPath, err)
			}
			c.logger.Debug("geckoterminal request",
				"method", method, "path", endpointPath, "status", resp.StatusCode,
				"duration", c.now().Sub(start), "attempt", attempt,
				"request_id", resp.Header.Get("x-request-id"))
			if key != "" {
				if ttl := cacheTTL(resp.Header, defaultCacheTTL); ttl > 0 {
					c.cache.Set(key, CacheEntry{
						Body:   body,
						Header: resp.Header.Clone(),
						ETag:   resp.Header.Get("ETag"),
					}, ttl)
				}
			}
			return body, resp.Header, nil
		}

		apiErr := c.parseAPIError(req, resp)
		_ = resp.Body.Close()
		if attempt < maxAttempts && c.retry.retryable(resp.StatusCode) {
			delay := c.retry.delay(attempt, apiErr.RetryAfter)
			c.logRetry(endpointPath, resp.StatusCode, attempt, delay, nil)
			if err := sleepContext(ctx, delay); err != nil {
				return nil, nil, err
			}
			continue
		}
		return nil, nil, apiErr
	}
}

// readLimited reads the body fully, enforcing the byte limit.
func readLimited(body io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("response body exceeds size limit")
	}
	return b, nil
}

// parseAPIError reads the (limited) error body and builds an *APIError.
// A non-JSON body is preserved in Body with Errors left empty.
func (c *Client) parseAPIError(req *http.Request, resp *http.Response) *APIError {
	body, _ := readLimited(resp.Body, maxErrorBodyBytes)
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Method:     req.Method,
		Path:       req.URL.Path,
		RequestID:  resp.Header.Get("x-request-id"),
		RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"), c.now()),
		Body:       body,
	}
	var env struct {
		Errors []ErrorDetail `json:"errors"`
		Meta   struct {
			RefID string `json:"ref_id"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &env); err == nil {
		apiErr.Errors = env.Errors
		apiErr.RefID = env.Meta.RefID
	}
	return apiErr
}

// logRetry emits a warning for a retryable failure.
func (c *Client) logRetry(path string, status, attempt int, delay time.Duration, cause error) {
	attrs := []any{
		"path", path,
		"attempt", attempt,
		"delay", delay,
	}
	if status != 0 {
		attrs = append(attrs, "status", status)
	}
	if cause != nil {
		attrs = append(attrs, "error", cause)
	}
	c.logger.Warn("geckoterminal retrying request", attrs...)
}

// fetchPage performs a GET request and decodes the JSON:API list
// envelope into a Page. page is the requested page number (1 when the
// caller did not set one).
func fetchPage[T any](ctx context.Context, c *Client, endpointPath string, q url.Values, page int) (*Page[T], error) {
	body, _, err := c.do(ctx, http.MethodGet, endpointPath, q)
	if err != nil {
		return nil, err
	}
	var p Page[T]
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("geckoterminal: decode %s: %w", endpointPath, err)
	}
	if page <= 0 {
		page = 1
	}
	p.Number = page
	p.HasNext = p.computeHasNext()
	return &p, nil
}

// fetchOne performs a GET request and decodes the JSON:API
// single-resource envelope into a Response.
func fetchOne[T any](ctx context.Context, c *Client, endpointPath string, q url.Values) (*Response[T], error) {
	body, _, err := c.do(ctx, http.MethodGet, endpointPath, q)
	if err != nil {
		return nil, err
	}
	var r Response[T]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("geckoterminal: decode %s: %w", endpointPath, err)
	}
	return &r, nil
}
