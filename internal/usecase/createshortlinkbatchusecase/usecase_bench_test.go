package createshortlinkbatchusecase

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/createshortlinkbatchusecase/mocks"
)

func BenchmarkCreateShortLinkBatchUsecase(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().AddBatch(context.Background(), []string{"https://example.com/1", "https://example.com/2"}, "").Return([]string{"abc0", "abc1"}, nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, "http://localhost:8080")
	
	reqBody := `[{"correlation_id": "1", "original_url": "https://example.com/1"}, {"correlation_id": "2", "original_url": "https://example.com/2"}]`
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		
		usecase.Execute(c)
	}
}

func BenchmarkCreateShortLinkBatchUsecaseWithUserID(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().AddBatch(context.Background(), []string{"https://example.com/1", "https://example.com/2"}, "user123").Return([]string{"abc0", "abc1"}, nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, "http://localhost:8080")
	
	reqBody := `[{"correlation_id": "1", "original_url": "https://example.com/1"}, {"correlation_id": "2", "original_url": "https://example.com/2"}]`
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", "user123")
		
		usecase.Execute(c)
	}
}