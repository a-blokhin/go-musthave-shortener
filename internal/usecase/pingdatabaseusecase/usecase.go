package pingdatabaseusecase

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Usecase struct {
	db     Database
	logger *zap.Logger
}

func New(db Database, logger *zap.Logger) *Usecase {
	return &Usecase{
		db:     db,
		logger: logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	if u.db == nil {
		u.logger.Info("No database configured, returning OK")
		c.Status(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := u.db.Ping(ctx); err != nil {
		u.logger.Error("Database ping failed", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	u.logger.Info("Database ping successful")
	c.Status(http.StatusOK)
}
