package geckoterminal_test

import (
	"encoding/json"
	"testing"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

// FuzzDecimalUnmarshal feeds arbitrary JSON values into Decimal. It must
// never panic, and MarshalJSON -> UnmarshalJSON must round-trip.
func FuzzDecimalUnmarshal(f *testing.F) {
	seeds := []string{
		`"1.00107146675173"`,
		`"716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726"`,
		`1723.98`,
		`0`,
		`-0.0000001`,
		`1e30`,
		`"abc"`,
		`""`,
		`null`,
		`true`,
		`[1,2]`,
		`{"a":1}`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		var d geckoterminal.Decimal
		if err := json.Unmarshal([]byte(input), &d); err != nil {
			t.Skip() // invalid JSON input is not interesting
		}
		b, err := d.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}
		var out geckoterminal.Decimal
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("round-trip unmarshal of %q: %v", b, err)
		}
		if out != d {
			t.Fatalf("round-trip changed %q -> %q", d, out)
		}
		_ = d.String()
		_ = d.IsZero()
		_, _ = d.Float64()
		_, _ = d.BigFloat()
	})
}

// FuzzCandleUnmarshal feeds arbitrary JSON into Candle. It must return
// an error or a fully decoded candle, never panic.
func FuzzCandleUnmarshal(f *testing.F) {
	seeds := []string{
		`[1712534400, 3454.61590249189, 3660.85954963415, 3417.91885296256, 3660.85954963415, 306823.277031161]`,
		`[0, 0, 0, 0, 0, 0]`,
		`[]`,
		`[1,2,3]`,
		`[1,2,3,4,5,6,7]`,
		`["x",1,2,3,4,5]`,
		`[1,"a","b","c","d","e"]`,
		`[1,null,null,null,null,null]`,
		`{}`,
		`123`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		var c geckoterminal.Candle
		err := json.Unmarshal([]byte(input), &c)
		if err != nil {
			return
		}
		b, err := c.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}
		var c2 geckoterminal.Candle
		if err := json.Unmarshal(b, &c2); err != nil {
			t.Fatalf("round-trip: %v", err)
		}
		if c2.Timestamp.Unix() != c.Timestamp.Unix() {
			t.Fatalf("timestamp changed: %v -> %v", c.Timestamp, c2.Timestamp)
		}
	})
}

// FuzzIncludedUnmarshal feeds arbitrary JSON into Included. Unknown
// resource types are ignored; decoding must never panic.
func FuzzIncludedUnmarshal(f *testing.F) {
	f.Add(`[]`)
	f.Add(`[{"id":"x","type":"token","attributes":{"name":"t","address":"a","symbol":"s","decimals":18,"image_url":"u"}}]`)
	f.Add(`[{"id":"d","type":"dex","attributes":{"name":"x"}},{"id":"n","type":"network","attributes":{"name":"y"}},{"id":"p","type":"pool","attributes":{"name":"z","address":"a","base_token_price_usd":"1","base_token_price_native_currency":"1","base_token_price_quote_token":"1","quote_token_price_usd":"1","quote_token_price_native_currency":"1","quote_token_price_base_token":"1","reserve_in_usd":"1"}},{"id":"m","type":"mystery"}]`)
	f.Add(`{}`)
	f.Add(`null`)
	f.Add(`[1,2,3]`)
	f.Add(`["a"]`)
	f.Fuzz(func(t *testing.T, input string) {
		var inc geckoterminal.Included
		if err := json.Unmarshal([]byte(input), &inc); err != nil {
			t.Skip()
		}
		for _, tok := range inc.Tokens {
			if got, ok := inc.Token(tok.ID); !ok || got.ID != tok.ID {
				t.Fatalf("token lookup failed for %q", tok.ID)
			}
		}
		for _, p := range inc.Pools {
			if got, ok := inc.Pool(p.ID); !ok || got.ID != p.ID {
				t.Fatalf("pool lookup failed for %q", p.ID)
			}
		}
		for _, d := range inc.Dexes {
			if got, ok := inc.DEX(d.ID); !ok || got.ID != d.ID {
				t.Fatalf("dex lookup failed for %q", d.ID)
			}
		}
	})
}
