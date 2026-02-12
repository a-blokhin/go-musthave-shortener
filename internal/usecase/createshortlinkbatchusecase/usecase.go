package createshortlinkbatchusecase

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/pkg/createshortlinkbatchpkg"
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
	var req createshortlinkbatchpkg.BatchRequest

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		u.logger.Error("Failed to decode JSON request",
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	defer c.Request.Body.Close()

	if len(req) == 0 {
		u.logger.Info("Empty batch provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Batch cannot be empty"})
		return
	}

	urls := make([]string, 0, len(req))
	for _, item := range req {
		if item.OriginalURL == "" {
			u.logger.Info("Empty URL in batch item",
				zap.String("correlation_id", item.CorrelationID))
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL cannot be empty"})
			return
		}
		urls = append(urls, item.OriginalURL)
	}

	aliases, err := u.linkRepo.AddBatch(urls)
	if err != nil {
		u.logger.Error("Failed to create short URLs",
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	response := make(createshortlinkbatchpkg.BatchResponse, 0, len(req))
	for i, item := range req {
		shortURL := u.baseURL + "/" + aliases[i]
		response = append(response, createshortlinkbatchpkg.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	c.JSON(http.StatusCreated, response)
}
