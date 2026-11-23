package geckoterminal_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

// fixture loads a golden file from testdata.
func fixture(t testing.TB, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name)) //nolint:gosec // testdata fixtures are trusted
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

// readFixture is like fixture but returns the error for benchmarks.
func readFixture(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join("testdata", name)) //nolint:gosec // testdata fixtures are trusted
}

// newTestClient creates a client against the given handler with the
// rate limiter and retries disabled.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*geckoterminal.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// fixtureHandler serves a single fixture with a 200 status.
func fixtureHandler(t *testing.T, name string, check func(r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(fixture(t, name)); err != nil {
			t.Errorf("write fixture: %v", err)
		}
	}
}

// statusHandler replies with the given status and body.
func statusHandler(status int, body string, headers map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestNewClientDefaults(t *testing.T) {
	t.Parallel()
	c, err := geckoterminal.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	for name, svc := range map[string]any{
		"Networks": c.Networks,
		"Dexes":    c.Dexes,
		"Pools":    c.Pools,
		"Tokens":   c.Tokens,
		"Trades":   c.Trades,
		"OHLCV":    c.OHLCV,
		"Simple":   c.Simple,
		"Search":   c.Search,
	} {
		if svc == nil {
			t.Errorf("service %s is nil", name)
		}
	}
}

func TestMustNewClientPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	geckoterminal.MustNewClient(geckoterminal.WithBaseURL(""))
}

func TestClientRequestHeaders(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json;version=20230203" {
			t.Errorf("Accept = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != geckoterminal.DefaultUserAgent {
			t.Errorf("User-Agent = %q", got)
		}
		if got := r.Header.Get("x-cg-pro-api-key"); got != "" {
			t.Errorf("unexpected API key header %q", got)
		}
		w.Write([]byte(`{"data":[]}`))
	})
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestClientBaseURLPathPrefix(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL+"/api/v2"),
		geckoterminal.WithRateLimit(0, 0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/api/v2/networks" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestClientCoinGeckoPro(t *testing.T) {
	t.Parallel()
	var gotPath, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("x-cg-pro-api-key")
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithCoinGeckoPro("test-key-123"),
		geckoterminal.WithBaseURL(srv.URL+"/api/v3/onchain"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotPath != "/api/v3/onchain/networks" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotKey != "test-key-123" {
		t.Fatalf("api key header = %q", gotKey)
	}
}

func TestClientOptionErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		opt  geckoterminal.Option
	}{
		{"empty base url", geckoterminal.WithBaseURL("")},
		{"bad scheme", geckoterminal.WithBaseURL("ftp://x")},
		{"no host", geckoterminal.WithBaseURL("https://")},
		{"empty pro key", geckoterminal.WithCoinGeckoPro("")},
		{"empty api key", geckoterminal.WithAPIKey("")},
		{"nil http client", geckoterminal.WithHTTPClient(nil)},
		{"zero timeout", geckoterminal.WithTimeout(0)},
		{"nil clock", geckoterminal.WithClock(nil)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := geckoterminal.NewClient(tc.opt); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestClientAPIVersionDisabled(t *testing.T) {
	t.Parallel()
	var accept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept = r.Header.Get("Accept")
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithAPIVersion(""),
		geckoterminal.WithRateLimit(0, 0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if accept != "application/json" {
		t.Fatalf("Accept = %q", accept)
	}
}

func TestClientCustomUserAgent(t *testing.T) {
	t.Parallel()
	var ua string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithUserAgent("my-app/2.0"),
		geckoterminal.WithRateLimit(0, 0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if ua != "my-app/2.0" {
		t.Fatalf("User-Agent = %q", ua)
	}
}

func TestClientRetryThenSuccess(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"errors":[{"status":"429","title":"Too Many Requests"}]}`))
			return
		}
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
			MaxRetries: 2,
			BaseDelay:  time.Millisecond,
			MaxDelay:   5 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestClientRetryExhausted(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`boom`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
			MaxRetries: 2,
			BaseDelay:  time.Millisecond,
			MaxDelay:   5 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Networks.List(context.Background(), nil)
	var apiErr *geckoterminal.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.StatusCode != 500 {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", calls.Load())
	}
}

func TestClientNoRetryOn404(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"status":"404","title":"Not Found"}],"meta":{"ref_id":"ref-1"}}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
			MaxRetries: 3,
			BaseDelay:  time.Millisecond,
			MaxDelay:   5 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Networks.List(context.Background(), nil)
	if !geckoterminal.IsNotFound(err) {
		t.Fatalf("err = %v, want IsNotFound", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestClientContextCanceled(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{"data":[]}`))
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Networks.List(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d, want 0", calls.Load())
	}
}

func TestClientRateLimiterEngages(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		// 6000/min = 100 tokens/sec: burst of 2 is instant, the third
		// request waits about 10ms.
		geckoterminal.WithRateLimit(6000, 2),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	start := time.Now()
	for i := 0; i < 3; i++ {
		if _, err := c.Networks.List(context.Background(), nil); err != nil {
			t.Fatalf("List %d: %v", i, err)
		}
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d", calls.Load())
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("rate-limited requests took too long: %v", elapsed)
	}
}

func TestClientRateLimiterContextCancel(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0.06, 1), // ~1 request per 1000s
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("first List: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = c.Networks.List(ctx, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
}

func TestClientMalformedJSON(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{not json`))
	})
	_, err := c.Networks.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected decode error")
	}
	var apiErr *geckoterminal.APIError
	if errors.As(err, &apiErr) {
		t.Fatalf("unexpected APIError: %v", err)
	}
}

func TestClientErrorNonJSONBody(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, statusHandler(http.StatusBadGateway,
		"<html><body>bad gateway</body></html>",
		map[string]string{"x-request-id": "req-42"}))
	_, err := c.Networks.List(context.Background(), nil)
	var apiErr *geckoterminal.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.RequestID != "req-42" {
		t.Fatalf("RequestID = %q", apiErr.RequestID)
	}
	if len(apiErr.Errors) != 0 {
		t.Fatalf("Errors = %+v", apiErr.Errors)
	}
	if len(apiErr.Body) == 0 {
		t.Fatal("Body is empty")
	}
}

func TestClientAPIErrorDetails(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, statusHandler(http.StatusNotFound, string(fixture(t, "error_404.json")),
		map[string]string{"x-request-id": "req-9"}))
	_, err := c.Networks.List(context.Background(), nil)
	var apiErr *geckoterminal.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v", err)
	}
	if apiErr.RefID != "b0ddc0cd-44c8-41ed-ad8c-a0692d443ca0" {
		t.Fatalf("RefID = %q", apiErr.RefID)
	}
	if len(apiErr.Errors) != 1 || apiErr.Errors[0].Status != "404" {
		t.Fatalf("Errors = %+v", apiErr.Errors)
	}
	if !geckoterminal.IsNotFound(err) {
		t.Fatal("IsNotFound = false")
	}
}

func TestClient429RetryAfter(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"errors":[{"status":"429","title":"Too Many Requests"}]}`))
			return
		}
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{
			MaxRetries: 1,
			BaseDelay:  time.Millisecond,
			MaxDelay:   5 * time.Millisecond,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d", calls.Load())
	}
}

func TestClient429IsRateLimited(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, statusHandler(http.StatusTooManyRequests,
		string(fixture(t, "error_429.json")), nil))
	_, err := c.Networks.List(context.Background(), nil)
	if !geckoterminal.IsRateLimited(err) {
		t.Fatalf("err = %v, want IsRateLimited", err)
	}
	var apiErr *geckoterminal.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v", err)
	}
}

func TestClientTimeoutOption(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithTimeout(30*time.Millisecond),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{MaxRetries: 0}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.Networks.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestClientCustomHTTPClient(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
		geckoterminal.WithTimeout(time.Nanosecond), // ignored: custom client wins
		geckoterminal.WithRateLimit(0, 0),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Networks.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestClientIncludedDecode(t *testing.T) {
	t.Parallel()
	c, _ := newTestClient(t, fixtureHandler(t, "pool_detail.json", nil))
	resp, err := c.Pools.Get(context.Background(), "eth", "0x60594a405d53811d3bc4766596efd80fd545a270", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(resp.Included.Tokens) != 2 {
		t.Fatalf("included tokens = %d", len(resp.Included.Tokens))
	}
	tok, ok := resp.Included.Token("eth_0x6b175474e89094c44da98b954eedeac495271d0f")
	if !ok || tok.Attributes.Symbol != "DAI" {
		t.Fatalf("included token lookup failed: %+v ok=%v", tok, ok)
	}
	if _, ok := resp.Included.DEX("uniswap_v3"); !ok {
		t.Fatal("included DEX lookup failed")
	}
	if _, ok := resp.Included.Network("eth"); !ok {
		t.Fatal("included network lookup failed")
	}
	if resp.Data.Relationships.BaseToken == nil {
		t.Fatal("base_token relationship missing")
	}
	if got := resp.Data.Relationships.BaseToken.ID(); got != "eth_0x6b175474e89094c44da98b954eedeac495271d0f" {
		t.Fatalf("base_token id = %q", got)
	}
}
