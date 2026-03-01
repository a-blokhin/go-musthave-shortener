package deleteurlsusecase

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/config"
)

type deleteRequest struct {
	shortURLs []string
	userID    string
}

type deleteJob struct {
	shortURLs []string
	userID    string
}

type Usecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger

	reqChan chan deleteRequest
	jobChan chan deleteJob

	batchSize     int
	flushInterval time.Duration
	workerCount   int

	stopOnce sync.Once
	stopChan chan struct{}

	wg sync.WaitGroup
}

func New(linkRepo LinkRepo, logger *zap.Logger, cfg config.DeleteURLsConfig) *Usecase {
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 256
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 100 * time.Millisecond
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}

	u := &Usecase{
		linkRepo:      linkRepo,
		logger:        logger,
		reqChan:       make(chan deleteRequest, cfg.BufferSize),
		jobChan:       make(chan deleteJob, cfg.BufferSize),
		batchSize:     cfg.BatchSize,
		flushInterval: cfg.FlushInterval,
		workerCount:   cfg.WorkerCount,
		stopChan:      make(chan struct{}),
	}

	u.wg.Add(1)
	go func() {
		defer u.wg.Done()
		u.batcher()
	}()

	for i := 0; i < u.workerCount; i++ {
		u.wg.Add(1)
		go func(workerID int) {
			defer u.wg.Done()
			u.worker(workerID)
		}(i)
	}

	u.logger.Info("DeleteURLs usecase started",
		zap.Int("workers", u.workerCount),
		zap.Int("batch_size", u.batchSize),
		zap.Duration("flush_interval", u.flushInterval),
		zap.Int("buffer_size", cfg.BufferSize),
	)

	return u
}

func (u *Usecase) Execute(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
		return
	}

	var shortURLs []string
	if err := c.ShouldBindJSON(&shortURLs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(shortURLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty list of URLs"})
		return
	}

	select {
	case <-u.stopChan:
		c.AbortWithStatus(http.StatusAccepted)
		return
	case u.reqChan <- deleteRequest{shortURLs: shortURLs, userID: userIDStr}:
		c.AbortWithStatus(http.StatusAccepted)
		return
	}
}

func (u *Usecase) Close() {
	u.stopOnce.Do(func() {
		close(u.stopChan)
		u.wg.Wait()
		u.logger.Info("DeleteURLs usecase stopped")
	})
}

func (u *Usecase) batcher() {
	ticker := time.NewTicker(u.flushInterval)
	defer ticker.Stop()

	pending := make(map[string][]string)

	flushUser := func(userID string, flushAll bool) {
		urls := pending[userID]
		if len(urls) == 0 {
			delete(pending, userID)
			return
		}

		for len(urls) > 0 {
			if !flushAll && len(urls) < u.batchSize {
				break
			}

			n := u.batchSize
			if n > len(urls) {
				n = len(urls)
			}

			chunk := make([]string, n)
			copy(chunk, urls[:n])

			select {
			case <-u.stopChan:

				return
			case u.jobChan <- deleteJob{shortURLs: chunk, userID: userID}:
			}

			urls = urls[n:]
		}

		if len(urls) == 0 {
			delete(pending, userID)
		} else {
			pending[userID] = urls
		}
	}

	flushAll := func() {
		for userID := range pending {
			flushUser(userID, true)
		}
	}

	for {
		select {
		case <-u.stopChan:
			flushAll()
			close(u.jobChan)
			return

		case <-ticker.C:
			flushAll()

		case req := <-u.reqChan:
			if req.userID == "" || len(req.shortURLs) == 0 {
				continue
			}
			pending[req.userID] = append(pending[req.userID], req.shortURLs...)
			flushUser(req.userID, false)
		}
	}
}

func (u *Usecase) worker(workerID int) {
	for job := range u.jobChan {
		if job.userID == "" || len(job.shortURLs) == 0 {
			continue
		}
		if err := u.linkRepo.BatchDelete(context.Background(), job.shortURLs, job.userID); err != nil {
			u.logger.Error("Failed to delete batch",
				zap.Int("worker_id", workerID),
				zap.String("user_id", job.userID),
				zap.Strings("short_urls", job.shortURLs),
				zap.Error(err),
			)
		}
	}
}
