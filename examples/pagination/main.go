// Command pagination demonstrates iterator-based pagination with All
// and Collect over the top pools of a network.
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, err := geckoterminal.NewClient()
	if err != nil {
		return err
	}

	// Iterate lazily: the next page is only fetched when needed.
	count := 0
	for pool, err := range client.Pools.AllTopByNetwork(ctx, "eth", nil) {
		if err != nil {
			return err
		}
		count++
		if count <= 5 {
			fmt.Println(pool.Attributes.Name)
		}
		if count >= 40 { // stop early; no more pages are fetched
			break
		}
	}
	fmt.Printf("iterated %d pools\n", count)

	// Or drain everything into a slice at once.
	all, err := geckoterminal.Collect(client.Dexes.AllByNetwork(ctx, "eth", nil))
	if err != nil {
		return err
	}
	fmt.Printf("collected %d dexes\n", len(all))
	return nil
}
