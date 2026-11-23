// Command pool-ohlcv prints daily candles for a pool.
//
// Set GECKOTERMINAL_POOL to a pool address on Ethereum mainnet; the
// DAI/WETH 0.05% Uniswap V3 pool is used by default.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	geckoterminal "github.com/tigusigalpa/geckoterminal-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	network := os.Getenv("GECKOTERMINAL_NETWORK")
	if network == "" {
		network = "eth"
	}
	pool := os.Getenv("GECKOTERMINAL_POOL")
	if pool == "" {
		pool = "0x60594a405d53811d3bc4766596efd80fd545a270"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := geckoterminal.NewClient()
	if err != nil {
		return err
	}

	resp, err := client.OHLCV.Get(ctx, network, pool, geckoterminal.TimeframeDay, &geckoterminal.OHLCVOptions{
		Limit: geckoterminal.Ptr(30),
	})
	if err != nil {
		return err
	}

	fmt.Printf("%s/%s daily candles\n", resp.Meta.Base.Symbol, resp.Meta.Quote.Symbol)
	for _, c := range resp.Data.Attributes.OHLCVList {
		fmt.Printf("%s  O=%s H=%s L=%s C=%s V=%s\n",
			c.Timestamp.Format("2006-01-02"),
			c.Open, c.High, c.Low, c.Close, c.Volume)
	}
	return nil
}
