package createshortlinkusecase

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/audit"
	"go-musthave-shortener/internal/middleware"
	"go-musthave-shortener/internal/model"
)

type Usecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
	baseURL  string
	audit    AuditEmitter
}

func New(linkRepo LinkRepo, logger *zap.Logger, baseURL string, audit AuditEmitter) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
		logger:   logger,
		baseURL:  baseURL,
		audit:    audit,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		u.logger.Error("Failed to read request body",
			zap.Error(err))
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}
	defer c.Request.Body.Close()

	url := strings.TrimSpace(string(body))
	if url == "" {
		u.logger.Info("Empty URL provided in request")
		c.String(http.StatusBadRequest, "Bad Request: URL is required")
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		userID = ""
	}

	alias, err := u.linkRepo.Add(c.Request.Context(), url, userID)
	if err != nil {
		var duplicateErr *model.DuplicateURLError
		if errors.As(err, &duplicateErr) {

			c.String(http.StatusConflict, u.baseURL+"/"+duplicateErr.ExistingShortURL)
			return
		}

		u.logger.Error("Failed to create short URL",
			zap.Error(err),
			zap.String("url", url))
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusCreated, u.baseURL+"/"+alias)

	if u.audit != nil {
		u.audit.Emit(audit.ActionShorten, userID, url)
	}
}
