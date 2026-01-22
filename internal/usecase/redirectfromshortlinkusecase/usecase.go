package redirectfromshortlinkusecase

import (
	"net/http"

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
		u.logger.Info("Empty alias parameter in redirect request")
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}

	
	originalURL, err := u.linkRepo.Get(alias)
	if err != nil {
		u.logger.Info("Failed to find original URL for alias", 
			zap.Error(err), 
			zap.String("alias", alias))
		c.String(http.StatusNotFound, "alias %q not found", alias)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}
