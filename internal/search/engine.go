package search

import "context"

type Engine interface {
	Search(ctx context.Context, q Query) ([]SearchResult, error)
}
