package getuserurlsusecase

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/middleware"
	"go-musthave-shortener/internal/repository"
)

type Usecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
	baseURL  string
}

func New(linkRepo LinkRepo, logger *zap.Logger, baseURL string) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
		logger:   logger,
		baseURL:  baseURL,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	userURLs, err := u.linkRepo.GetByUserID(c.Request.Context(), userID)
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
		expectedResponse, err := url.JoinPath(u.baseURL, userURL.ShortURL)
		if err != nil {
			u.logger.Error("Failed to create expected response", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		response[i] = repository.UserURL{
			ShortURL:    expectedResponse,
			OriginalURL: userURL.OriginalURL,
		}
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}
