package createshortlinkbatchusecase

import (
	"go-musthave-shortener/internal/auth"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

	userID, err := auth.GetUserIDFromGinContext(c)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	if len(aliases) != len(req) {
		u.logger.Error("Mismatch between request count and aliases count",
			zap.Int("request_count", len(req)),
			zap.Int("aliases_count", len(aliases)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	resp := make([]ShortenBatchResponse, len(req))
	for i, item := range req {
		expectedResponse, err := url.JoinPath(u.baseURL, aliases[i])
		if err != nil {
			u.logger.Error("Failed to create expected response", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		resp[i] = ShortenBatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      expectedResponse,
		}
	}

	c.JSON(http.StatusCreated, resp)
}
