package getstatsusecase

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Usecase struct {
	repo   StatsRepository
	logger *zap.Logger
}

func New(repo StatsRepository, logger *zap.Logger) *Usecase {
	return &Usecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	ctx := context.Background()

	stats, err := u.repo.GetStats(ctx)
	if err != nil {
		u.logger.Error("Failed to get stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
