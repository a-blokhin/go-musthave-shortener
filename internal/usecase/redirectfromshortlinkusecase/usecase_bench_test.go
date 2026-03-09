package redirectfromshortlinkusecase

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase/mocks"
)

func BenchmarkRedirectFromShortLinkUsecase(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().Get(context.Background(), "abc123").Return("https://example.com", nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, nil)
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/abc123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "abc123"}}
		
		usecase.Execute(c)
	}
}

func BenchmarkRedirectFromShortLinkUsecaseWithUserID(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := mocks.NewLinkRepo(b)
	mockRepo.EXPECT().Get(context.Background(), "abc123").Return("https://example.com", nil)
	
	logger := zap.NewNop()
	usecase := New(mockRepo, logger, nil)
	
	
	for b.Loop() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/abc123", nil)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "abc123"}}
		c.Set("userID", "user123")
		
		usecase.Execute(c)
	}
}