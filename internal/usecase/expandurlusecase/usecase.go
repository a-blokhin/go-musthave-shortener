package expandurlusecase

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go-musthave-shortener/internal/audit"
	"go-musthave-shortener/internal/model"
)

type ExpandURLUsecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
	audit    AuditEmitter
}

func New(linkRepo LinkRepo, logger *zap.Logger, audit AuditEmitter) *ExpandURLUsecase {
	return &ExpandURLUsecase{
		linkRepo: linkRepo,
		logger:   logger,
		audit:    audit,
	}
}

func (u *ExpandURLUsecase) Execute(ctx context.Context, alias string, userID string) (string, error) {
	if alias == "" {
		u.logger.Info("Empty alias provided in request")
		return "", errors.New("ID is required")
	}

	originalURL, err := u.linkRepo.Get(ctx, alias)
	if err != nil {
		u.logger.Info("URL not found for alias",
			zap.String("alias", alias),
			zap.Error(err))

		var deletedErr *model.DeletedURLError
		if errors.As(err, &deletedErr) {
			return "", errors.New("URL has been deleted")
		}

		return "", errors.New("URL not found")
	}

	if u.audit != nil {
		u.audit.Emit(audit.ActionFollow, userID, originalURL)
	}

	return originalURL, nil
}
