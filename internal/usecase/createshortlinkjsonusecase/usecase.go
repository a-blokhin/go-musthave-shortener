package createshortlinkjsonusecase

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

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
	
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		u.logger.Error("Failed to decode JSON request", 
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	defer c.Request.Body.Close()

	if req.URL == "" {
		u.logger.Info("Empty URL provided in JSON request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	alias, err := u.linkRepo.Add(req.URL)
	if err != nil {
		u.logger.Error("Failed to create short URL", 
			zap.Error(err), 
			zap.String("url", req.URL))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	shortURL := u.baseURL + "/" + alias
	resp := createshortlinkjsonpkg.Response{
		Result: shortURL,
	}

	c.JSON(http.StatusCreated, resp)
}