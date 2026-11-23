package geckoterminal

import (
	"encoding/json"
	"errors"
)

// ResourceIdentifier is a JSON:API resource identifier.
type ResourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Relationship is a JSON:API relationship. Data is either a single
// resource identifier object or an array of them.
type Relationship struct {
	Data json.RawMessage `json:"data,omitempty"`
}

// ID returns the related resource ID, or "" when data is an array,
// null or absent.
func (r Relationship) ID() string {
	var v ResourceIdentifier
	if err := json.Unmarshal(r.Data, &v); err != nil {
		return ""
	}
	return v.ID
}

// IDs returns the related resource IDs when data is an array, or nil
// otherwise.
func (r Relationship) IDs() []string {
	var vs []ResourceIdentifier
	if err := json.Unmarshal(r.Data, &vs); err != nil {
		return nil
	}
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.ID
	}
	return out
}

// Links holds JSON:API pagination links. Only GET /networks currently
// returns them.
type Links struct {
	First *string `json:"first,omitempty"`
	Prev  *string `json:"prev,omitempty"`
	Next  *string `json:"next,omitempty"`
	Last  *string `json:"last,omitempty"`
}

// Included holds the heterogeneous resources from a JSON:API "included"
// array, dispatched by resource type. Unknown types are ignored.
type Included struct {
	Tokens   []Token
	Pools    []Pool
	Dexes    []DEX
	Networks []Network
}

// UnmarshalJSON dispatches each included resource by its "type" field.
func (inc *Included) UnmarshalJSON(b []byte) error {
	var raws []json.RawMessage
	if err := json.Unmarshal(b, &raws); err != nil {
		return err
	}
	for _, raw := range raws {
		var head struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &head); err != nil {
			return err
		}
		switch head.Type {
		case "token":
			var v Token
			if err := json.Unmarshal(raw, &v); err != nil {
				return err
			}
			inc.Tokens = append(inc.Tokens, v)
		case "pool":
			var v Pool
			if err := json.Unmarshal(raw, &v); err != nil {
				return err
			}
			inc.Pools = append(inc.Pools, v)
		case "dex":
			var v DEX
			if err := json.Unmarshal(raw, &v); err != nil {
				return err
			}
			inc.Dexes = append(inc.Dexes, v)
		case "network":
			var v Network
			if err := json.Unmarshal(raw, &v); err != nil {
				return err
			}
			inc.Networks = append(inc.Networks, v)
		}
	}
	return nil
}

// Token returns an included token by ID.
func (inc Included) Token(id string) (Token, bool) {
	for _, v := range inc.Tokens {
		if v.ID == id {
			return v, true
		}
	}
	return Token{}, false
}

// Pool returns an included pool by ID.
func (inc Included) Pool(id string) (Pool, bool) {
	for _, v := range inc.Pools {
		if v.ID == id {
			return v, true
		}
	}
	return Pool{}, false
}

// DEX returns an included DEX by ID.
func (inc Included) DEX(id string) (DEX, bool) {
	for _, v := range inc.Dexes {
		if v.ID == id {
			return v, true
		}
	}
	return DEX{}, false
}

// Network returns an included network by ID.
func (inc Included) Network(id string) (Network, bool) {
	for _, v := range inc.Networks {
		if v.ID == id {
			return v, true
		}
	}
	return Network{}, false
}

// Response is a single-resource JSON:API response.
type Response[T any] struct {
	Data     T
	Included Included
	Links    *Links
	Meta     json.RawMessage
}

// UnmarshalJSON decodes the JSON:API envelope. The "data" member is
// required.
func (r *Response[T]) UnmarshalJSON(b []byte) error {
	var env struct {
		Data     json.RawMessage `json:"data"`
		Included Included        `json:"included"`
		Links    *Links          `json:"links"`
		Meta     json.RawMessage `json:"meta"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return err
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return errors.New("geckoterminal: response has no data")
	}
	var data T
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return err
	}
	r.Data = data
	r.Included = env.Included
	r.Links = env.Links
	r.Meta = env.Meta
	return nil
}

// Page is a paginated JSON:API list response.
type Page[T any] struct {
	Items    []T
	Included Included
	Links    *Links
	Meta     json.RawMessage
	// Number is the requested page number, 1 when unset.
	Number int
	// HasNext reports whether another page is expected.
	HasNext bool
}

// UnmarshalJSON decodes the JSON:API list envelope.
func (p *Page[T]) UnmarshalJSON(b []byte) error {
	var env struct {
		Data     []T             `json:"data"`
		Included Included        `json:"included"`
		Links    *Links          `json:"links"`
		Meta     json.RawMessage `json:"meta"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return err
	}
	p.Items = env.Data
	p.Included = env.Included
	p.Links = env.Links
	p.Meta = env.Meta
	return nil
}

// NextCursor returns meta.next_cursor for cursor-paginated endpoints
// (CoinGecko Pro trades). It returns "" when the field is absent.
func (p *Page[T]) NextCursor() string {
	if len(p.Meta) == 0 {
		return ""
	}
	var m struct {
		NextCursor string `json:"next_cursor"`
	}
	if err := json.Unmarshal(p.Meta, &m); err != nil {
		return ""
	}
	return m.NextCursor
}

// computeHasNext derives HasNext after decoding. Explicit links win,
// then a non-empty cursor, then the page-based heuristic.
func (p *Page[T]) computeHasNext() bool {
	if p.Links != nil {
		return p.Links.Next != nil
	}
	if p.NextCursor() != "" {
		return true
	}
	return p.Number < maxPage && len(p.Items) > 0
}

// Interval is a metrics time window used by volume, price change and
// transaction attributes.
type Interval string

// Metrics intervals returned by the API.
const (
	IntervalM5  Interval = "m5"
	IntervalM15 Interval = "m15"
	IntervalM30 Interval = "m30"
	IntervalH1  Interval = "h1"
	IntervalH6  Interval = "h6"
	IntervalH24 Interval = "h24"
)

// String returns the API representation of the interval.
func (i Interval) String() string { return string(i) }

// IntervalValues maps an interval to a decimal value. It is a map
// (rather than a fixed struct) so that new intervals added by the API
// do not break decoding.
type IntervalValues map[Interval]Decimal

// Get returns the value for the interval, or "" when absent.
func (v IntervalValues) Get(i Interval) Decimal { return v[i] }

// M5 returns the 5-minute value.
func (v IntervalValues) M5() Decimal { return v[IntervalM5] }

// M15 returns the 15-minute value.
func (v IntervalValues) M15() Decimal { return v[IntervalM15] }

// M30 returns the 30-minute value.
func (v IntervalValues) M30() Decimal { return v[IntervalM30] }

// H1 returns the 1-hour value.
func (v IntervalValues) H1() Decimal { return v[IntervalH1] }

// H6 returns the 6-hour value.
func (v IntervalValues) H6() Decimal { return v[IntervalH6] }

// H24 returns the 24-hour value.
func (v IntervalValues) H24() Decimal { return v[IntervalH24] }

// TransactionStats is buy/sell activity for one interval.
type TransactionStats struct {
	Buys          int `json:"buys"`
	Sells         int `json:"sells"`
	Buyers        int `json:"buyers"`
	Sellers       int `json:"sellers"`
	UniqueTraders int `json:"unique_traders"`
}

// Transactions maps an interval to transaction stats.
type Transactions map[Interval]TransactionStats

// Get returns the stats for the interval, or the zero value when absent.
func (t Transactions) Get(i Interval) TransactionStats { return t[i] }

// H24 returns the 24-hour stats.
func (t Transactions) H24() TransactionStats { return t[IntervalH24] }

// H6 returns the 6-hour stats.
func (t Transactions) H6() TransactionStats { return t[IntervalH6] }

// H1 returns the 1-hour stats.
func (t Transactions) H1() TransactionStats { return t[IntervalH1] }

// M5 returns the 5-minute stats.
func (t Transactions) M5() TransactionStats { return t[IntervalM5] }

// Network is a supported blockchain network.
type Network struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Attributes NetworkAttributes `json:"attributes"`
}

// NetworkAttributes holds network attributes.
type NetworkAttributes struct {
	Name                     string `json:"name"`
	CoinGeckoAssetPlatformID string `json:"coingecko_asset_platform_id"`
}

// DEX is a decentralized exchange.
type DEX struct {
	ID         string        `json:"id"`
	Type       string        `json:"type"`
	Attributes DEXAttributes `json:"attributes"`
}

// DEXAttributes holds DEX attributes.
type DEXAttributes struct {
	Name string `json:"name"`
}
