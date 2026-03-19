package getuserurlsusecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/auth"
	"go-musthave-shortener/internal/repository"
	"go-musthave-shortener/internal/usecase/getuserurlsusecasegeneric"
)

type Usecase struct {
	getUserURLsUsecase *getuserurlsusecasegeneric.GetUserURLsUsecase
	logger             *zap.Logger
}

func New(getUserURLsUsecase *getuserurlsusecasegeneric.GetUserURLsUsecase, logger *zap.Logger) *Usecase {
	return &Usecase{
		getUserURLsUsecase: getUserURLsUsecase,
		logger:             logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	userID, err := auth.GetUserIDFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	userURLs, err := u.getUserURLsUsecase.Execute(c.Request.Context(), userID)
	if err != nil {
		u.logger.Error("Failed to get user URLs",
			zap.Error(err),
			zap.String("userID", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	u.logger.Info("Retrieved user URLs", zap.Int("count", len(userURLs)), zap.String("userID", userID))

	if len(userURLs) == 0 {
		u.logger.Info("User has no URLs, returning 204")
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	u.logger.Info("User has URLs, returning 200")

	response := make([]repository.UserURL, len(userURLs))
	for i, userURL := range userURLs {
		response[i] = repository.UserURL{
			ShortURL:    userURL.ShortURL,
			OriginalURL: userURL.OriginalURL,
		}
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}
