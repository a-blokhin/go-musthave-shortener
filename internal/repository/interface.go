package repository

import "context"

// UserURL represents the response format for user URLs
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type LinkRepository interface {
	Add(ctx context.Context, url string, userID string) (string, error)
	AddBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	Get(ctx context.Context, alias string) (string, error)
	GetByUserID(ctx context.Context, userID string) ([]UserURL, error)
}
