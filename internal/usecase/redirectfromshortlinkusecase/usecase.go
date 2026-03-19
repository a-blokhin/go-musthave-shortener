package redirectfromshortlinkusecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/auth"
	"go-musthave-shortener/internal/usecase/expandurlusecase"
)

type Usecase struct {
	expandURLUsecase *expandurlusecase.ExpandURLUsecase
	logger           *zap.Logger
}

func New(expandURLUsecase *expandurlusecase.ExpandURLUsecase, logger *zap.Logger) *Usecase {
	return &Usecase{
		expandURLUsecase: expandURLUsecase,
		logger:           logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	alias := c.Param("id")
	if alias == "" {
		u.logger.Info("Empty alias provided in request")
		c.String(http.StatusBadRequest, "Bad Request: alias is required")
		return
	}

	userID, err := auth.GetUserIDFromGinContext(c)
	if err != nil {
		userID = ""
	}

	originalURL, err := u.expandURLUsecase.Execute(c.Request.Context(), alias, userID)
	if err != nil {
		u.logger.Info("URL not found for alias",
			zap.String("alias", alias),
			zap.Error(err))
		c.String(http.StatusNotFound, "Not Found")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}
