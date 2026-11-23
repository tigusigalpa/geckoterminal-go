package geckoterminal

import (
	"context"
	"iter"
)

// All iterates over pages 1..10, yielding each item with a nil error.
// fetch is called with the page number and must return the page.
//
// Iteration stops when the consumer returns early, the context is done,
// fetch returns an error, Page.HasNext is false, or the hard cap of 10
// pages is reached. On an error it yields the zero value and the error,
// then stops.
func All[T any](ctx context.Context, fetch func(context.Context, int) (*Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for page := 1; page <= maxPage; page++ {
			if err := ctx.Err(); err != nil {
				var zero T
				yield(zero, err)
				return
			}
			p, err := fetch(ctx, page)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			if p == nil {
				return
			}
			for _, item := range p.Items {
				if !yield(item, nil) {
					return
				}
			}
			if !p.HasNext {
				return
			}
		}
	}
}

// Collect drains an iterator into a slice. It returns the items
// collected before the first error, together with that error.
func Collect[T any](seq iter.Seq2[T, error]) ([]T, error) {
	var out []T
	for v, err := range seq {
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
	return out, nil
}
