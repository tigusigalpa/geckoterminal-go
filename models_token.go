package geckoterminal

import "time"

// Token is an on-chain token with market data.
type Token struct {
	ID            string              `json:"id"`
	Type          string              `json:"type"`
	Attributes    TokenAttributes     `json:"attributes"`
	Relationships *TokenRelationships `json:"relationships,omitempty"`
}

// TokenRelationships holds token relationships.
type TokenRelationships struct {
	TopPools *Relationship `json:"top_pools,omitempty"`
}

// TokenAttributes holds token attributes.
type TokenAttributes struct {
	Name                  string         `json:"name"`
	Address               string         `json:"address"`
	Symbol                string         `json:"symbol"`
	Decimals              int            `json:"decimals"`
	ImageURL              string         `json:"image_url"`
	BannerImageURL        *string        `json:"banner_image_url,omitempty"`
	TotalSupply           Decimal        `json:"total_supply,omitempty"`
	NormalizedTotalSupply Decimal        `json:"normalized_total_supply,omitempty"`
	CoinGeckoCoinID       *string        `json:"coingecko_coin_id,omitempty"`
	PriceUSD              Decimal        `json:"price_usd,omitempty"`
	FDVUSD                *Decimal       `json:"fdv_usd,omitempty"`
	TotalReserveInUSD     Decimal        `json:"total_reserve_in_usd,omitempty"`
	VolumeUSD             IntervalValues `json:"volume_usd,omitempty"`
	MarketCapUSD          *Decimal       `json:"market_cap_usd,omitempty"`
}

// TokenInfo is extended token metadata returned by the token info,
// pool tokens info and recently updated tokens endpoints.
type TokenInfo struct {
	ID            string                  `json:"id"`
	Type          string                  `json:"type"`
	Attributes    TokenInfoAttributes     `json:"attributes"`
	Relationships *TokenInfoRelationships `json:"relationships,omitempty"`
}

// TokenInfoRelationships holds token info relationships.
type TokenInfoRelationships struct {
	Network *Relationship `json:"network,omitempty"`
}

// TokenInfoAttributes holds token info attributes. Note the pair of
// category-ID fields: the token info endpoint emits "gt_categories_id"
// while tokens/info_recently_updated emits "gt_category_ids". Both are
// real API keys.
type TokenInfoAttributes struct {
	Address                    string          `json:"address"`
	Name                       string          `json:"name"`
	Symbol                     string          `json:"symbol"`
	Decimals                   int             `json:"decimals,omitempty"`
	ImageURL                   string          `json:"image_url"`
	Image                      *TokenImage     `json:"image,omitempty"`
	BannerImageURL             *string         `json:"banner_image_url,omitempty"`
	CoinGeckoCoinID            *string         `json:"coingecko_coin_id,omitempty"`
	Websites                   []string        `json:"websites,omitempty"`
	Description                string          `json:"description,omitempty"`
	GTScore                    float64         `json:"gt_score,omitempty"`
	GTScoreDetails             *GTScoreDetails `json:"gt_score_details,omitempty"`
	DiscordURL                 *string         `json:"discord_url,omitempty"`
	TelegramHandle             *string         `json:"telegram_handle,omitempty"`
	TwitterHandle              *string         `json:"twitter_handle,omitempty"`
	Categories                 []string        `json:"categories,omitempty"`
	GTCategoriesID             []string        `json:"gt_categories_id,omitempty"`
	GTCategoryIDs              []string        `json:"gt_category_ids,omitempty"`
	Holders                    *TokenHolders   `json:"holders,omitempty"`
	MintAuthority              *string         `json:"mint_authority,omitempty"`
	FreezeAuthority            *string         `json:"freeze_authority,omitempty"`
	IsHoneypot                 *bool           `json:"is_honeypot,omitempty"`
	DeveloperAddress           *string         `json:"developer_address,omitempty"`
	DeveloperHoldingPercentage *float64        `json:"developer_holding_percentage,omitempty"`
	MetadataUpdatedAt          *time.Time      `json:"metadata_updated_at,omitempty"`
}

// TokenImage holds CoinGecko-hosted token images at fixed sizes.
type TokenImage struct {
	Thumb string `json:"thumb"`
	Small string `json:"small"`
	Large string `json:"large"`
}

// GTScoreDetails is the GeckoTerminal score breakdown.
type GTScoreDetails struct {
	Pool        float64 `json:"pool"`
	Transaction float64 `json:"transaction"`
	Creation    float64 `json:"creation"`
	Info        float64 `json:"info"`
	Holders     float64 `json:"holders"`
}

// TokenHolders summarizes holder statistics for a token.
type TokenHolders struct {
	Count                  int64              `json:"count"`
	DistributionPercentage HolderDistribution `json:"distribution_percentage"`
	LastUpdated            *time.Time         `json:"last_updated,omitempty"`
}

// HolderDistribution is the holder concentration breakdown.
type HolderDistribution struct {
	Top10   Decimal `json:"top_10"`
	P11To30 Decimal `json:"11_30"`
	P31To50 Decimal `json:"31_50"`
	Rest    Decimal `json:"rest"`
}

// TokenPrices is the flattened attribute payload of the simple token
// price endpoint. Maps are keyed by token address.
type TokenPrices struct {
	TokenPrices              map[string]Decimal `json:"token_prices"`
	MarketCapUSD             map[string]Decimal `json:"market_cap_usd,omitempty"`
	H24VolumeUSD             map[string]Decimal `json:"h24_volume_usd,omitempty"`
	H24PriceChangePercentage map[string]Decimal `json:"h24_price_change_percentage,omitempty"`
	TotalReserveInUSD        map[string]Decimal `json:"total_reserve_in_usd,omitempty"`
}
