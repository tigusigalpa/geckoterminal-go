package geckoterminal

const (
	// Namespace is the human-readable project namespace.
	Namespace = "Tigusigalpa/GeckoTerminal"
	// Version is the semantic version of the SDK.
	Version = "0.1.0"
	// DefaultUserAgent is sent with every request unless overridden.
	DefaultUserAgent = "geckoterminal-go/" + Version + " (" + Namespace + "; +https://github.com/tigusigalpa/geckoterminal-go )"
)

const (
	// DefaultBaseURL is the public GeckoTerminal API v2 base URL.
	DefaultBaseURL = "https://api.geckoterminal.com/api/v2"
	// ProBaseURL is the CoinGecko Pro on-chain API base URL used by
	// WithCoinGeckoPro. It already contains the "/onchain" path prefix.
	ProBaseURL = "https://pro-api.coingecko.com/api/v3/onchain"
	// DefaultAPIVersion is the GeckoTerminal API version sent in the
	// Accept header as "application/json;version=<DefaultAPIVersion>".
	DefaultAPIVersion = "20230203"
)
