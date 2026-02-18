package createshortlinkbatchusecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/middleware"
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

type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (u *Usecase) Execute(c *gin.Context) {
	var req []ShortenBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		u.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	if len(req) == 0 {
		u.logger.Info("Empty batch request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Batch request cannot be empty"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		userID = ""
	}

	urls := make([]string, len(req))
	for i, item := range req {
		if item.OriginalURL == "" {
			u.logger.Info("Empty URL in batch item", zap.String("correlation_id", item.CorrelationID))
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL cannot be empty"})
			return
		}
		urls[i] = item.OriginalURL
	}

	aliases, err := u.linkRepo.AddBatch(c.Request.Context(), urls, userID)
	if err != nil {
		u.logger.Error("Failed to create short URLs in batch", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if len(aliases) != len(req) {
		u.logger.Error("Mismatch between request count and aliases count",
			zap.Int("request_count", len(req)),
			zap.Int("aliases_count", len(aliases)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	resp := make([]ShortenBatchResponse, len(req))
	for i, item := range req {
		resp[i] = ShortenBatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      u.baseURL + "/" + aliases[i],
		}
	}

	c.JSON(http.StatusCreated, resp)
}
