package geckoterminal

import (
	"encoding/json"
	"fmt"
	"time"
)

// Timeframe is an OHLCV timeframe path parameter.
type Timeframe string

// OHLCV timeframes.
const (
	TimeframeDay    Timeframe = "day"
	TimeframeHour   Timeframe = "hour"
	TimeframeMinute Timeframe = "minute"
)

// String returns the API representation of the timeframe.
func (t Timeframe) String() string { return string(t) }

// Valid reports whether the timeframe is supported.
func (t Timeframe) Valid() bool {
	switch t {
	case TimeframeDay, TimeframeHour, TimeframeMinute:
		return true
	}
	return false
}

// Aggregate is an OHLCV aggregation factor. Valid values depend on the
// timeframe: day accepts 1; hour accepts 1, 4, 12; minute accepts
// 1, 5, 15.
type Aggregate string

// OHLCV aggregation factors.
const (
	Aggregate1  Aggregate = "1"
	Aggregate4  Aggregate = "4"
	Aggregate5  Aggregate = "5"
	Aggregate12 Aggregate = "12"
	Aggregate15 Aggregate = "15"
)

// String returns the API representation of the aggregation factor.
func (a Aggregate) String() string { return string(a) }

// validFor reports whether the aggregation factor is allowed for the
// given timeframe.
func (a Aggregate) validFor(tf Timeframe) bool {
	switch tf {
	case TimeframeDay:
		return a == Aggregate1
	case TimeframeHour:
		return a == Aggregate1 || a == Aggregate4 || a == Aggregate12
	case TimeframeMinute:
		return a == Aggregate1 || a == Aggregate5 || a == Aggregate15
	}
	return false
}

// Currency is the OHLCV price currency.
type Currency string

// OHLCV price currencies.
const (
	CurrencyUSD   Currency = "usd"
	CurrencyToken Currency = "token"
)

// String returns the API representation of the currency.
func (c Currency) String() string { return string(c) }

var validCurrencies = map[Currency]bool{
	CurrencyUSD: true, CurrencyToken: true,
}

// Candle is a single OHLCV candle decoded from the array form
// [timestamp, open, high, low, close, volume].
type Candle struct {
	Timestamp time.Time
	Open      Decimal
	High      Decimal
	Low       Decimal
	Close     Decimal
	Volume    Decimal
}

// UnmarshalJSON decodes [timestamp, open, high, low, close, volume],
// where timestamp is a Unix second and the remaining elements are
// numbers.
func (c *Candle) UnmarshalJSON(b []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if len(raw) != 6 {
		return fmt.Errorf("geckoterminal: invalid candle length %d", len(raw))
	}
	var ts int64
	if err := json.Unmarshal(raw[0], &ts); err != nil {
		return fmt.Errorf("geckoterminal: invalid candle timestamp: %w", err)
	}
	c.Timestamp = time.Unix(ts, 0).UTC()
	fields := []*Decimal{&c.Open, &c.High, &c.Low, &c.Close, &c.Volume}
	for i, dst := range fields {
		if err := dst.UnmarshalJSON(raw[i+1]); err != nil {
			return fmt.Errorf("geckoterminal: invalid candle value: %w", err)
		}
	}
	return nil
}

// MarshalJSON encodes the candle back into the array form.
func (c Candle) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		c.Timestamp.Unix(),
		c.Open, c.High, c.Low, c.Close, c.Volume,
	})
}

// OHLCVResponse is the typed OHLCV response with base/quote token
// metadata.
type OHLCVResponse struct {
	Data OHLCVData
	Meta OHLCVMeta
}

// OHLCVData is the OHLCV resource object.
type OHLCVData struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes OHLCVAttributes `json:"attributes"`
}

// OHLCVAttributes holds the candle list.
type OHLCVAttributes struct {
	OHLCVList []Candle `json:"ohlcv_list"`
}

// OHLCVMeta describes the base and quote tokens of the chart.
type OHLCVMeta struct {
	Base  TokenMeta `json:"base"`
	Quote TokenMeta `json:"quote"`
}

// TokenMeta is the abbreviated token description returned in OHLCV
// metadata.
type TokenMeta struct {
	Address         string  `json:"address"`
	Name            string  `json:"name"`
	Symbol          string  `json:"symbol"`
	CoinGeckoCoinID *string `json:"coingecko_coin_id,omitempty"`
}
