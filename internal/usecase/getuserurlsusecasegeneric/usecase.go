package getuserurlsusecasegeneric

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go-musthave-shortener/internal/repository"
)

type GetUserURLsUsecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
	baseURL  string
}

func New(linkRepo LinkRepo, logger *zap.Logger, baseURL string) *GetUserURLsUsecase {
	return &GetUserURLsUsecase{
		linkRepo: linkRepo,
		logger:   logger,
		baseURL:  baseURL,
	}
}

func (u *GetUserURLsUsecase) Execute(ctx context.Context, userID string) ([]repository.UserURL, error) {
	userURLs, err := u.linkRepo.GetByUserID(ctx, userID)
	if err != nil {
		u.logger.Error("Failed to get user URLs",
			zap.Error(err),
			zap.String("userID", userID))
		return nil, errors.New("failed to get user URLs")
	}

	u.logger.Info("Retrieved user URLs", zap.Int("count", len(userURLs)), zap.String("userID", userID))

	if len(userURLs) == 0 {
		return nil, nil
	}

	response := make([]repository.UserURL, len(userURLs))
	for i, userURL := range userURLs {
		response[i] = repository.UserURL{
			ShortURL:    u.baseURL + "/" + userURL.ShortURL,
			OriginalURL: userURL.OriginalURL,
		}
	}

	return response, nil
}
