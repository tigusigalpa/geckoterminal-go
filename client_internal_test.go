package geckoterminal

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestReadLimited(t *testing.T) {
	t.Parallel()
	b, err := readLimited(strings.NewReader("hello"), 10)
	if err != nil || string(b) != "hello" {
		t.Fatalf("readLimited = %q, %v", b, err)
	}
	if _, err := readLimited(strings.NewReader("0123456789abcdef"), 10); err == nil {
		t.Fatal("expected size limit error")
	}
	if _, err := readLimited(strings.NewReader("0123456789"), 10); err != nil {
		t.Fatalf("at-limit read failed: %v", err)
	}
}

func TestBuildURL(t *testing.T) {
	t.Parallel()
	c := MustNewClient()
	got := c.buildURL("/networks/eth/pools/multi/0x1,0x2", nil)
	if !strings.HasSuffix(got, "/api/v2/networks/eth/pools/multi/0x1,0x2") {
		t.Fatalf("url = %q", got)
	}
	if c.acceptHeader() != "application/json;version=20230203" {
		t.Fatalf("accept = %q", c.acceptHeader())
	}
	c2 := MustNewClient(WithAPIVersion(""))
	if c2.acceptHeader() != "application/json" {
		t.Fatalf("accept = %q", c2.acceptHeader())
	}
}

func TestJoinPathEsc(t *testing.T) {
	t.Parallel()
	if got := joinPath("networks", "eth", "pools"); got != "/networks/eth/pools" {
		t.Fatalf("joinPath = %q", got)
	}
	if got := esc("a b/c?"); got != "a%20b%2Fc%3F" {
		t.Fatalf("esc = %q", got)
	}
}

func TestDecimalUnmarshalDirect(t *testing.T) {
	t.Parallel()
	var d Decimal
	// Raw JSON values that are neither string nor null are taken verbatim.
	if err := d.UnmarshalJSON([]byte(`{"a":1}`)); err != nil {
		t.Fatalf("object: %v", err)
	}
	if d != `{"a":1}` {
		t.Fatalf("d = %q", d)
	}
	// A malformed quoted string hits the inner unmarshal error.
	if err := d.UnmarshalJSON([]byte(`"bad`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestSleepContextNegative(t *testing.T) {
	t.Parallel()
	if err := sleepContext(context.Background(), -time.Second); err != nil {
		t.Fatalf("err = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleepContext(ctx, -time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestAggregateValidForEdge(t *testing.T) {
	t.Parallel()
	if Aggregate("7").validFor(TimeframeDay) {
		t.Fatal("7 valid for day")
	}
	if Aggregate1.validFor(Timeframe("week")) {
		t.Fatal("aggregate valid for unknown timeframe")
	}
}

func TestCandleUnmarshalEdges(t *testing.T) {
	t.Parallel()
	var c Candle
	if err := c.UnmarshalJSON([]byte(`[]`)); err == nil {
		t.Fatal("empty array should error")
	}
	if err := c.UnmarshalJSON([]byte(`["x",1,2,3,4,5]`)); err == nil {
		t.Fatal("bad timestamp should error")
	}
	if err := c.UnmarshalJSON([]byte(`[1,2,3]`)); err == nil {
		t.Fatal("short array should error")
	}
	if err := c.UnmarshalJSON([]byte(`[1712534400,"a","b","c","d","e"]`)); err != nil {
		t.Fatalf("string fields: %v", err)
	}
	if c.Open != "a" || c.Volume != "e" {
		t.Fatalf("candle = %+v", c)
	}
	// Round-trip.
	b, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var c2 Candle
	if err := c2.UnmarshalJSON(b); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
}

func TestMemoryCacheEdgeCases(t *testing.T) {
	t.Parallel()
	if NewMemoryCache(0).capacity != 256 {
		t.Fatal("default capacity")
	}
	if NewMemoryCache(-3).capacity != 256 {
		t.Fatal("negative capacity")
	}
	c := NewMemoryCache(2)
	c.Set("k", CacheEntry{Body: []byte("v1")}, time.Minute)
	c.Set("k", CacheEntry{Body: []byte("v2")}, time.Minute) // update existing
	got, _ := c.Get("k")
	if string(got.Body) != "v2" {
		t.Fatalf("body = %q", got.Body)
	}
}

func TestIncludedUnmarshalErrors(t *testing.T) {
	t.Parallel()
	var inc Included
	if err := inc.UnmarshalJSON([]byte(`{`)); err == nil {
		t.Fatal("malformed should error")
	}
	if err := inc.UnmarshalJSON([]byte(`[123]`)); err == nil {
		t.Fatal("non-object element should error")
	}
	if err := inc.UnmarshalJSON([]byte(`[{"type":"token","id":"x","attributes":"nope"}]`)); err == nil {
		t.Fatal("bad token attributes should error")
	}
	if err := inc.UnmarshalJSON([]byte(`[{"type":"pool","id":"x","attributes":"nope"}]`)); err == nil {
		t.Fatal("bad pool attributes should error")
	}
	if err := inc.UnmarshalJSON([]byte(`[{"type":"dex","id":"x","attributes":"nope"}]`)); err == nil {
		t.Fatal("bad dex attributes should error")
	}
	if err := inc.UnmarshalJSON([]byte(`[{"type":"network","id":"x","attributes":"nope"}]`)); err == nil {
		t.Fatal("bad network attributes should error")
	}
}

func TestResponseUnmarshalEdges(t *testing.T) {
	t.Parallel()
	var r Response[Network]
	if err := json.Unmarshal([]byte(`{"data":{"id":"eth","type":"network","attributes":{"name":"Ethereum","coingecko_asset_platform_id":"ethereum"}},"links":{"next":"http://x"},"meta":{"a":1}}`), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r.Links == nil || r.Links.Next == nil || *r.Links.Next != "http://x" {
		t.Fatalf("links = %+v", r.Links)
	}
	if string(r.Meta) != `{"a":1}` {
		t.Fatalf("meta = %s", r.Meta)
	}
}
