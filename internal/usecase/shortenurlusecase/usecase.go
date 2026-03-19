package shortenurlusecase

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"go-musthave-shortener/internal/audit"
	"go-musthave-shortener/internal/model"
)

type ShortenURLUsecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
	baseURL  string
	audit    AuditEmitter
}

func New(linkRepo LinkRepo, logger *zap.Logger, baseURL string, audit AuditEmitter) *ShortenURLUsecase {
	return &ShortenURLUsecase{
		linkRepo: linkRepo,
		logger:   logger,
		baseURL:  baseURL,
		audit:    audit,
	}
}

func (u *ShortenURLUsecase) Execute(ctx context.Context, urlStr string, userID string) (string, error) {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		u.logger.Info("Empty URL provided in request")
		return "", errors.New("URL is required")
	}

	alias, err := u.linkRepo.Add(ctx, urlStr, userID)
	if err != nil {
		var duplicateErr *model.DuplicateURLError
		if errors.As(err, &duplicateErr) {
			expResp, err := url.JoinPath(u.baseURL, duplicateErr.ExistingShortURL)
			if err != nil {
				u.logger.Error("Failed to create response", zap.Error(err))
				return "", errors.New("failed to create short URL")
			}
			return expResp, nil
		}

		u.logger.Error("Failed to create short URL",
			zap.Error(err),
			zap.String("url", urlStr))
		return "", errors.New("failed to create short URL")
	}

	expResp, err := url.JoinPath(u.baseURL, alias)
	if err != nil {
		u.logger.Error("Failed to create response", zap.Error(err))
		return "", errors.New("failed to create short URL")
	}

	if u.audit != nil {
		u.audit.Emit(audit.ActionShorten, userID, urlStr)
	}

	return expResp, nil
}
