package getstatsusecase

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/getstatsusecasegeneric"
)

type Usecase struct {
	getStatsUsecase *getstatsusecasegeneric.GetStatsUsecase
	logger          *zap.Logger
}

func NewHTTPHandler(getStatsUsecase *getstatsusecasegeneric.GetStatsUsecase, logger *zap.Logger) *Usecase {
	return &Usecase{
		getStatsUsecase: getStatsUsecase,
		logger:          logger,
	}
}

func (h *Usecase) Execute(c *gin.Context) {
	ctx := context.Background()

	stats, err := h.getStatsUsecase.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to get stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
