package geckoterminal

import (
	"encoding/json"
	"errors"
	"math/big"
	"testing"
)

func TestDecimalUnmarshal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  Decimal
	}{
		{"string", `"1.00107146675173"`, "1.00107146675173"},
		{"number", `1723.98`, "1723.98"},
		{"integer", `42`, "42"},
		{"negative", `"-4716.5197581542"`, "-4716.5197581542"},
		{"null", `null`, ""},
		{"empty string", `""`, ""},
		{"high precision", `"716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726"`,
			"716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d Decimal
			if err := json.Unmarshal([]byte(tc.input), &d); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if d != tc.want {
				t.Fatalf("got %q, want %q", d, tc.want)
			}
		})
	}
}

func TestDecimalRoundTrip(t *testing.T) {
	t.Parallel()
	in := Decimal("716191484.679024586262538640101397833507485654499768471382550941923442594835115853914415164260594726")
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out Decimal
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out != in {
		t.Fatalf("round trip changed value: %q -> %q", in, out)
	}
}

func TestDecimalFloat64(t *testing.T) {
	t.Parallel()
	f, err := Decimal("1.5").Float64()
	if err != nil || f != 1.5 {
		t.Fatalf("Float64 = %v, %v", f, err)
	}
	if _, err := Decimal("").Float64(); !errors.Is(err, ErrEmptyDecimal) {
		t.Fatalf("empty Float64 err = %v", err)
	}
	if _, err := Decimal("notanumber").Float64(); err == nil {
		t.Fatal("expected parse error")
	}
	if got := Decimal("3.25").MustFloat64(); got != 3.25 {
		t.Fatalf("MustFloat64 = %v", got)
	}
}

func TestDecimalMustFloat64Panics(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	Decimal("").MustFloat64()
}

func TestDecimalBigFloat(t *testing.T) {
	t.Parallel()
	big1, ok := Decimal("0.1").BigFloat()
	if !ok {
		t.Fatal("BigFloat failed")
	}
	want, _ := new(big.Float).SetString("0.1")
	if big1.Cmp(want) != 0 {
		t.Fatalf("BigFloat = %v", big1)
	}
	if _, ok := Decimal("").BigFloat(); ok {
		t.Fatal("empty BigFloat should fail")
	}
	if _, ok := Decimal("garbage").BigFloat(); ok {
		t.Fatal("garbage BigFloat should fail")
	}
}

func TestDecimalIsZero(t *testing.T) {
	t.Parallel()
	for _, d := range []Decimal{"", "0", "0.0", "0.00000", "-0.0"} {
		if !d.IsZero() {
			t.Fatalf("%q IsZero = false", d)
		}
	}
	for _, d := range []Decimal{"1", "0.1", "-2.5"} {
		if d.IsZero() {
			t.Fatalf("%q IsZero = true", d)
		}
	}
	// A non-numeric string is treated as non-zero.
	if Decimal("abc").IsZero() {
		t.Fatal("abc IsZero = true")
	}
}

func TestDecimalInStruct(t *testing.T) {
	t.Parallel()
	var a struct {
		V Decimal `json:"v"`
	}
	if err := json.Unmarshal([]byte(`{"v":"123.456789012345678901234567890"}`), &a); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if a.V != "123.456789012345678901234567890" {
		t.Fatalf("v = %q", a.V)
	}
}
