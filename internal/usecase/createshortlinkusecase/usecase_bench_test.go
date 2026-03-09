package createshortlinkusecase

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/createshortlinkusecase/mocks"
)

func BenchmarkCreateShortLinkUsecase(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().Add(context.Background(), "https://example.com", "").Return("abc123", nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, "http://localhost:8080", nil)
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString("https://example.com"))
		
		usecase.Execute(c)
	}
}

func BenchmarkCreateShortLinkUsecaseWithUserID(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().Add(context.Background(), "https://example.com", "user123").Return("abc123", nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, "http://localhost:8080", nil)
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString("https://example.com"))
		c.Set("userID", "user123")
		
		usecase.Execute(c)
	}
}