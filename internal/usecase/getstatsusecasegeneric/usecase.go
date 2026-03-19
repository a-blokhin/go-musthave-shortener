package getstatsusecasegeneric

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go-musthave-shortener/internal/repository"
)

type GetStatsUsecase struct {
	repo   StatsRepository
	logger *zap.Logger
}

func New(repo StatsRepository, logger *zap.Logger) *GetStatsUsecase {
	return &GetStatsUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *GetStatsUsecase) Execute(ctx context.Context) (repository.Stats, error) {
	stats, err := u.repo.GetStats(ctx)
	if err != nil {
		u.logger.Error("Failed to get stats", zap.Error(err))
		return repository.Stats{}, errors.New("failed to get stats")
	}

	return stats, nil
}
