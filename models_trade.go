package geckoterminal

import "time"

// TradeKind is the direction of a trade.
type TradeKind string

// Trade directions.
const (
	TradeKindBuy  TradeKind = "buy"
	TradeKindSell TradeKind = "sell"
)

// String returns the API representation of the trade kind.
func (k TradeKind) String() string { return string(k) }

// Trade is a single DEX trade in a pool.
type Trade struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes TradeAttributes `json:"attributes"`
}

// TradeAttributes holds trade attributes.
type TradeAttributes struct {
	BlockNumber              int64     `json:"block_number"`
	BlockTimestamp           time.Time `json:"block_timestamp"`
	TxHash                   string    `json:"tx_hash"`
	TxFromAddress            string    `json:"tx_from_address"`
	Kind                     TradeKind `json:"kind"`
	FromTokenAddress         string    `json:"from_token_address"`
	ToTokenAddress           string    `json:"to_token_address"`
	FromTokenAmount          Decimal   `json:"from_token_amount"`
	ToTokenAmount            Decimal   `json:"to_token_amount"`
	PriceFromInUSD           Decimal   `json:"price_from_in_usd"`
	PriceToInUSD             Decimal   `json:"price_to_in_usd"`
	PriceFromInCurrencyToken Decimal   `json:"price_from_in_currency_token"`
	PriceToInCurrencyToken   Decimal   `json:"price_to_in_currency_token"`
	VolumeInUSD              Decimal   `json:"volume_in_usd"`
	IsAnomalous              bool      `json:"is_anomalous"`
}
