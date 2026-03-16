package getstatsusecase

import (
	"context"

	"go-musthave-shortener/internal/repository"
)

type StatsRepository interface {
	GetStats(ctx context.Context) (repository.Stats, error)
}
