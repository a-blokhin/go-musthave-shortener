package createshortlinkjsonusecase

import (
	"errors"
	"net/http"
	"net/url"
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

	reqUrl := strings.TrimSpace(req.URL)
	if reqUrl == "" {
		u.logger.Info("Empty URL provided in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		userID = ""
	}

	alias, err := u.linkRepo.Add(c.Request.Context(), reqUrl, userID)
	if err != nil {
		var duplicateErr *model.DuplicateURLError
		if errors.As(err, &duplicateErr) {

			expResp, err := url.JoinPath(u.baseURL, duplicateErr.ExistingShortURL)
			if err != nil {
				u.logger.Error("Failed to create expected response", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
				return
			}

			resp := createshortlinkjsonpkg.Response{
				Result: expResp,
			}
			c.JSON(http.StatusConflict, resp)
			return
		}

		u.logger.Error("Failed to create short URL",
			zap.Error(err),
			zap.String("url", reqUrl))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	expResp, err := url.JoinPath(u.baseURL, alias)
	if err != nil {
		u.logger.Error("Failed to create expected response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	resp := createshortlinkjsonpkg.Response{
		Result: expResp,
	}
	c.JSON(http.StatusCreated, resp)
}
