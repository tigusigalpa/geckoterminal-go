package geckoterminal_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func BenchmarkDecimalFloat64(b *testing.B) {
	d := geckoterminal.Decimal("716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Float64(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecimalBigFloat(b *testing.B) {
	d := geckoterminal.Decimal("716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := d.BigFloat(); !ok {
			b.Fatal("BigFloat failed")
		}
	}
}

func BenchmarkPoolUnmarshal(b *testing.B) {
	data, err := readFixture("pool_detail.json")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var r geckoterminal.Response[geckoterminal.Pool]
		if err := json.Unmarshal(data, &r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOHLCVUnmarshal(b *testing.B) {
	// Build a 1000-candle payload.
	var sb strings.Builder
	sb.WriteString(`{"data":{"id":"x","type":"ohlcv_request_response","attributes":{"ohlcv_list":[`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, "[%d,1.5,2.5,1.0,2.0,12345.678]", 1712534400+int64(i)*86400)
	}
	sb.WriteString(`]}},"meta":{"base":{"address":"a","name":"n","symbol":"s"},"quote":{"address":"q","name":"n","symbol":"s"}}}`)
	data := []byte(sb.String())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var r geckoterminal.OHLCVResponse
		if err := json.Unmarshal(data, &r); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientPoolsTrending(b *testing.B) {
	data, err := readFixture("pools_trending.json")
	if err != nil {
		b.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer srv.Close()
	c, err := geckoterminal.NewClient(
		geckoterminal.WithBaseURL(srv.URL),
		geckoterminal.WithRateLimit(0, 0),
		geckoterminal.WithRetryPolicy(geckoterminal.RetryPolicy{MaxRetries: 0}),
	)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Pools.Trending(ctx, nil); err != nil {
			b.Fatal(err)
		}
	}
}
