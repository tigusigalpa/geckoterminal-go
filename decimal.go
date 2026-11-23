package geckoterminal

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
)

// ErrEmptyDecimal is returned when converting an empty Decimal to a
// number.
var ErrEmptyDecimal = errors.New("geckoterminal: empty decimal")

// Decimal is an exact decimal value represented as its original string
// form. The API returns monetary values as strings, sometimes with 60+
// digits after the decimal point; Decimal preserves them verbatim and
// never loses precision.
//
// Use Float64 for display or loose math and BigFloat for precise
// arithmetic. Decimal accepts JSON strings, numbers and null.
type Decimal string

// String returns the original string representation.
func (d Decimal) String() string { return string(d) }

// IsZero reports whether the decimal is empty or numerically zero.
func (d Decimal) IsZero() bool {
	if d == "" {
		return true
	}
	f, ok := new(big.Float).SetString(string(d))
	return ok && f.Sign() == 0
}

// Float64 converts the decimal to float64. Use only for display or
// loose math: precision beyond float64 is discarded. It returns
// ErrEmptyDecimal for an empty decimal.
func (d Decimal) Float64() (float64, error) {
	if d == "" {
		return 0, ErrEmptyDecimal
	}
	return strconv.ParseFloat(string(d), 64)
}

// MustFloat64 is like Float64 but panics on error.
func (d Decimal) MustFloat64() float64 {
	f, err := d.Float64()
	if err != nil {
		panic(err)
	}
	return f
}

// BigFloat converts the decimal to an arbitrary-precision big.Float.
// The boolean result reports whether the conversion succeeded; an empty
// decimal returns (nil, false).
func (d Decimal) BigFloat() (*big.Float, bool) {
	if d == "" {
		return nil, false
	}
	f, ok := new(big.Float).SetString(string(d))
	return f, ok
}

// MarshalJSON encodes the decimal as a JSON string.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}

// UnmarshalJSON accepts a JSON string, number, or null. Null decodes to
// the empty decimal.
func (d *Decimal) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*d = Decimal(s)
		return nil
	}
	*d = Decimal(string(b))
	return nil
}
