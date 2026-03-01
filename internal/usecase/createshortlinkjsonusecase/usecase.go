package createshortlinkjsonusecase

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/middleware"
	"go-musthave-shortener/internal/model"
	"go-musthave-shortener/pkg/createshortlinkjsonpkg"
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
	var req createshortlinkjsonpkg.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		u.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	url := strings.TrimSpace(req.URL)
	if url == "" {
		u.logger.Info("Empty URL provided in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		userID = ""
	}

	alias, err := u.linkRepo.Add(c.Request.Context(), url, userID)
	if err != nil {
		var duplicateErr *model.DuplicateURLError
		if errors.As(err, &duplicateErr) {

			resp := createshortlinkjsonpkg.Response{
				Result: u.baseURL + "/" + duplicateErr.ExistingShortURL,
			}
			c.JSON(http.StatusConflict, resp)
			return
		}

		u.logger.Error("Failed to create short URL",
			zap.Error(err),
			zap.String("url", url))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	resp := createshortlinkjsonpkg.Response{
		Result: u.baseURL + "/" + alias,
	}
	c.JSON(http.StatusCreated, resp)
}
