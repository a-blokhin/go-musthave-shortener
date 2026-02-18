package deleteurlsusecase

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type deleteRequest struct {
	shortURLs []string
	userID    string
}

type Usecase struct {
	linkRepo      LinkRepo
	logger        *zap.Logger
	reqChan       chan deleteRequest
	closeChan     chan struct{}
	batchSize     int
	flushInterval time.Duration
}

func New(linkRepo LinkRepo, logger *zap.Logger) *Usecase {
	u := &Usecase{
		linkRepo:      linkRepo,
		logger:        logger,
		reqChan:       make(chan deleteRequest, 512),
		closeChan:     make(chan struct{}),
		batchSize:     256,
		flushInterval: 200 * time.Millisecond,
	}

	for i := 0; i < 2; i++ {
		go u.process()
	}

	return u
}

func (u *Usecase) Execute(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		u.logger.Info("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		u.logger.Info("Invalid user ID type in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var shortURLs []string
	if err := c.ShouldBindJSON(&shortURLs); err != nil {
		u.logger.Info("Failed to parse request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if len(shortURLs) == 0 {
		u.logger.Info("Empty list of URLs provided for deletion")
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty list of URLs"})
		return
	}

	go u.enqueue(shortURLs, userIDStr)

	c.Status(http.StatusAccepted)
	c.Writer.WriteHeaderNow()
}

func (u *Usecase) Close() {
	select {
	case <-u.closeChan:
		return
	default:
		close(u.closeChan)
	}
}

func (u *Usecase) enqueue(shortURLs []string, userID string) {
	select {
	case <-u.closeChan:
		return
	case u.reqChan <- deleteRequest{
		shortURLs: shortURLs,
		userID:    userID,
	}:
	}
}

func (u *Usecase) process() {
	ticker := time.NewTicker(u.flushInterval)
	defer ticker.Stop()

	pending := make(map[string][]string)

	flushUser := func(userID string) {
		urls := pending[userID]
		for len(urls) > 0 {
			n := u.batchSize
			if n > len(urls) {
				n = len(urls)
			}

			chunk := urls[:n]

			if err := u.linkRepo.BatchDelete(context.Background(), chunk, userID); err != nil {
				u.logger.Error(
					"Failed to delete batch",
					zap.Strings("short_urls", chunk),
					zap.String("user_id", userID),
					zap.Error(err),
				)
			}

			urls = urls[n:]
		}

		delete(pending, userID)
	}

	flushAll := func() {
		for userID := range pending {
			flushUser(userID)
		}
	}

	for {
		select {
		case req := <-u.reqChan:
			if len(req.shortURLs) == 0 || req.userID == "" {
				continue
			}

			pending[req.userID] = append(pending[req.userID], req.shortURLs...)

			for len(pending[req.userID]) >= u.batchSize {
				urls := pending[req.userID]
				chunk := urls[:u.batchSize]

				if err := u.linkRepo.BatchDelete(context.Background(), chunk, req.userID); err != nil {
					u.logger.Error(
						"Failed to delete batch",
						zap.Strings("short_urls", chunk),
						zap.String("user_id", req.userID),
						zap.Error(err),
					)
				}

				pending[req.userID] = urls[u.batchSize:]
				if len(pending[req.userID]) == 0 {
					delete(pending, req.userID)
					break
				}
			}

		case <-ticker.C:
			flushAll()

		case <-u.closeChan:
			flushAll()
			u.logger.Info("Close delete worker")
			return
		}
	}
}
