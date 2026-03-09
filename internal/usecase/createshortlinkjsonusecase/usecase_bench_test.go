package createshortlinkjsonusecase

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase/mocks"
)

func BenchmarkCreateShortLinkJSONUsecase(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().Add(context.Background(), "https://example.com", "").Return("abc123", nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, "http://localhost:8080", nil)
	
	reqBody := `{"url": "https://example.com"}`
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/shorten", bytes.NewBufferString(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		
		usecase.Execute(c)
	}
}

func BenchmarkCreateShortLinkJSONUsecaseWithUserID(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().Add(context.Background(), "https://example.com", "user123").Return("abc123", nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, "http://localhost:8080", nil)
	
	reqBody := `{"url": "https://example.com"}`
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/shorten", bytes.NewBufferString(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("userID", "user123")
		
		usecase.Execute(c)
	}
}