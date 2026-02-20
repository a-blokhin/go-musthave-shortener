package repository

import "context"

type LinkRepository interface {
	Add(ctx context.Context, url string) (string, error)
	AddBatch(ctx context.Context, urls []string) ([]string, error)
	Get(ctx context.Context, alias string) (string, error)
}
