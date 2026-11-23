// Command trending prints trending pools with included base tokens
// resolved through JSON:API relationships.
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

	pools, err := client.Pools.Trending(ctx, &geckoterminal.TrendingPoolsOptions{
		Duration: geckoterminal.Ptr(geckoterminal.Duration24h),
		Include: []geckoterminal.Include{
			geckoterminal.IncludeBaseToken,
			geckoterminal.IncludeDEX,
		},
	})
	if err != nil {
		return err
	}

	for _, p := range pools.Items {
		base := ""
		if p.Relationships.BaseToken != nil {
			if tok, ok := pools.Included.Token(p.Relationships.BaseToken.ID()); ok {
				base = tok.Attributes.Symbol
			}
		}
		fmt.Printf("%-40s base=%-10s price=%s USD\n",
			p.Attributes.Name, base, p.Attributes.BaseTokenPriceUSD)
	}
	return nil
}
