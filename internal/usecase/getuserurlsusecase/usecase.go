package getuserurlsusecase

import (
	"net/http"

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
	userID, exists := middleware.GetUserID(c)
	if !exists {
		u.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userURLs, err := u.linkRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		u.logger.Error("Failed to get user URLs",
			zap.Error(err),
			zap.String("userID", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
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
	for i, url := range userURLs {
		response[i] = repository.UserURL{
			ShortURL:    u.baseURL + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}
