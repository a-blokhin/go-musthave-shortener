package getstatsusecasegeneric

import (
	"context"

	"go-musthave-shortener/internal/repository"
)

//go:generate mockery --name=StatsRepository --output=./mocks --outpkg=mocks --filename=stats_repository_mock.go --with-expecter
type StatsRepository interface {
	GetStats(ctx context.Context) (repository.Stats, error)
}
