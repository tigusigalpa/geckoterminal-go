// Command token-price prints USD prices for a set of tokens using the
// simple price endpoint.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := geckoterminal.NewClient()
	if err != nil {
		return err
	}

	resp, err := client.Simple.TokenPrices(ctx, "eth", []string{
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
		"0xdac17f958d2ee523a2206206994597c13d831ec7", // USDT
		"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // USDC
	}, &geckoterminal.TokenPricesOptions{
		IncludeMarketCap: geckoterminal.Ptr(true),
	})
	if err != nil {
		return err
	}

	for addr, price := range resp.Data.TokenPrices {
		mcap := ""
		if v, ok := resp.Data.MarketCapUSD[addr]; ok {
			mcap = " mcap=" + v.String()
		}
		fmt.Printf("%s  $%s%s\n", addr, price, mcap)
	}
	return nil
}
