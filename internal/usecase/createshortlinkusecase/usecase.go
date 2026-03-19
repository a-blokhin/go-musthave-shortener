package createshortlinkusecase

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/auth"
	"go-musthave-shortener/internal/usecase/shortenurlusecase"
)

type Usecase struct {
	shortenURLUsecase *shortenurlusecase.ShortenURLUsecase
	logger            *zap.Logger
}

func New(shortenURLUsecase *shortenurlusecase.ShortenURLUsecase, logger *zap.Logger) *Usecase {
	return &Usecase{
		shortenURLUsecase: shortenURLUsecase,
		logger:            logger,
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

	userID, err := auth.GetUserIDFromGinContext(c)
	if err != nil {
		userID = ""
	}

	result, err := u.shortenURLUsecase.Execute(c.Request.Context(), url, userID)
	if err != nil {
		u.logger.Error("Failed to create short URL",
			zap.Error(err),
			zap.String("url", url))
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusCreated, result)
}
