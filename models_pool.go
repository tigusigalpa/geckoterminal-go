package geckoterminal

import "time"

// Pool is a DEX liquidity pool. It is returned by the pool list,
// detail, multi, search and token-pool endpoints, and may appear inside
// JSON:API "included".
type Pool struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    PoolAttributes    `json:"attributes"`
	Relationships PoolRelationships `json:"relationships,omitempty"`
}

// PoolRelationships holds pool relationships.
type PoolRelationships struct {
	BaseToken  *Relationship `json:"base_token,omitempty"`
	QuoteToken *Relationship `json:"quote_token,omitempty"`
	DEX        *Relationship `json:"dex,omitempty"`
	Network    *Relationship `json:"network,omitempty"`
}

// PoolAttributes holds pool attributes. The same shape serves the
// list resource, the pool detail resource and included pools, so fields
// absent from a given payload keep their zero value. Pointer fields may
// be null in API responses.
type PoolAttributes struct {
	Name                            string         `json:"name"`
	PoolName                        string         `json:"pool_name,omitempty"`
	Address                         string         `json:"address"`
	PoolFeePercentage               Decimal        `json:"pool_fee_percentage,omitempty"`
	BaseTokenPriceUSD               Decimal        `json:"base_token_price_usd"`
	BaseTokenPriceNativeCurrency    Decimal        `json:"base_token_price_native_currency"`
	BaseTokenBalance                Decimal        `json:"base_token_balance,omitempty"`
	BaseTokenLiquidityUSD           Decimal        `json:"base_token_liquidity_usd,omitempty"`
	QuoteTokenPriceUSD              Decimal        `json:"quote_token_price_usd"`
	QuoteTokenPriceNativeCurrency   Decimal        `json:"quote_token_price_native_currency"`
	QuoteTokenBalance               Decimal        `json:"quote_token_balance,omitempty"`
	QuoteTokenLiquidityUSD          Decimal        `json:"quote_token_liquidity_usd,omitempty"`
	BaseTokenPriceQuoteToken        Decimal        `json:"base_token_price_quote_token"`
	QuoteTokenPriceBaseToken        Decimal        `json:"quote_token_price_base_token"`
	PoolCreatedAt                   *time.Time     `json:"pool_created_at,omitempty"`
	FDVUSD                          *Decimal       `json:"fdv_usd,omitempty"`
	MarketCapUSD                    *Decimal       `json:"market_cap_usd,omitempty"`
	TokenPriceUSD                   Decimal        `json:"token_price_usd,omitempty"`
	PriceChangePercentage           IntervalValues `json:"price_change_percentage,omitempty"`
	Transactions                    Transactions   `json:"transactions,omitempty"`
	VolumeUSD                       IntervalValues `json:"volume_usd,omitempty"`
	NetBuyVolumeUSD                 IntervalValues `json:"net_buy_volume_usd,omitempty"`
	BuyVolumeUSD                    IntervalValues `json:"buy_volume_usd,omitempty"`
	SellVolumeUSD                   IntervalValues `json:"sell_volume_usd,omitempty"`
	ReserveInUSD                    Decimal        `json:"reserve_in_usd"`
	LockedLiquidityPercentage       *Decimal       `json:"locked_liquidity_percentage,omitempty"`
	SentimentVotePositivePercentage *float64       `json:"sentiment_vote_positive_percentage,omitempty"`
	SentimentVoteNegativePercentage *float64       `json:"sentiment_vote_negative_percentage,omitempty"`
	CommunitySusReport              *int           `json:"community_sus_report,omitempty"`
	// LaunchpadDetails is undocumented in the API and therefore kept
	// as an unstructured object.
	LaunchpadDetails map[string]any `json:"launchpad_details,omitempty"`
}
