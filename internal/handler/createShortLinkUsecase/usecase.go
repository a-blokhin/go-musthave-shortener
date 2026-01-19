package createshortlinkusecase

import (
	"io"
	"net/http"
	"strings"

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

	
	alias, err := u.linkRepo.Add(url)
	if err != nil {
		u.logger.Error("Failed to create short URL", 
			zap.Error(err), 
			zap.String("url", url))
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusCreated, u.baseURL+"/"+alias)
}
