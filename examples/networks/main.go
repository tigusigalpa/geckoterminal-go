// Command networks lists the supported blockchain networks and walks
// all pages via the All iterator.
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

	page, err := client.Networks.List(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Printf("page %d: %d networks (has_next=%v)\n", page.Number, len(page.Items), page.HasNext)

	fmt.Println("all networks:")
	for n, err := range client.Networks.All(ctx, nil) {
		if err != nil {
			return err
		}
		fmt.Printf("  %-20s %s\n", n.ID, n.Attributes.Name)
	}
	return nil
}
