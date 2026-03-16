package grpcgetuserurlsusecase

import (
	"context"

	"go-musthave-shortener/internal/repository"
)

type LinkRepo interface {
	GetByUserID(ctx context.Context, userID string) ([]repository.UserURL, error)
}

type UserURL struct {
	ShortURL    string
	OriginalURL string
}
