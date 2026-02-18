package redirectfromshortlinkusecase

import (
	"errors"
	"net/http"

	"go-musthave-shortener/internal/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Usecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
}

func New(linkRepo LinkRepo, logger *zap.Logger) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
		logger:   logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	alias := c.Param("id")
	if alias == "" {
		u.logger.Info("Empty alias provided in request")
		c.String(http.StatusBadRequest, "Bad Request: alias is required")
		return
	}

	originalURL, err := u.linkRepo.Get(c.Request.Context(), alias)
	if err != nil {
		u.logger.Info("URL not found for alias",
			zap.String("alias", alias),
			zap.Error(err))

		if errors.Is(err, &model.DeletedURLError{}) {
			c.String(http.StatusGone, "Gone")
			return
		}

		c.String(http.StatusNotFound, "Not Found")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}
