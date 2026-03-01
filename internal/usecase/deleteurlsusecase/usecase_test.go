package deleteurlsusecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/deleteurlsusecase/mocks"
)

func TestDeleteURLs_Success(t *testing.T) {
	userID := "test-user-id"
	shortURLs := []string{"abc123", "def456", "ghi789"}

	mockRepo := mocks.NewLinkRepo(t)

	var wg sync.WaitGroup
	wg.Add(1)
	mockRepo.EXPECT().BatchDelete(mock.Anything, shortURLs, userID).Return(nil).Run(func(ctx context.Context, urls []string, uid string) {
		wg.Done()
	})

	usecase := New(mockRepo, zap.NewNop(), DefaultConfig())
	defer usecase.Close()

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer([]byte(`["abc123","def456","ghi789"]`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	wg.Wait()
}

func TestDeleteURLs_EmptyList(t *testing.T) {
	userID := "test-user-id"

	mockRepo := mocks.NewLinkRepo(t)

	usecase := New(mockRepo, zap.NewNop(), DefaultConfig())
	defer usecase.Close()

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer([]byte(`[]`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "empty list of URLs" {
		t.Errorf("expected error message 'empty list of URLs', got %s", response["error"])
	}
}

func TestDeleteURLs_InvalidJSON(t *testing.T) {
	userID := "test-user-id"

	mockRepo := mocks.NewLinkRepo(t)

	usecase := New(mockRepo, zap.NewNop(), DefaultConfig())
	defer usecase.Close()

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "invalid request body" {
		t.Errorf("expected error message 'invalid request body', got %s", response["error"])
	}
}

func TestDeleteURLs_NoUserID(t *testing.T) {
	mockRepo := mocks.NewLinkRepo(t)

	usecase := New(mockRepo, zap.NewNop(), DefaultConfig())
	defer usecase.Close()

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer([]byte(`["abc123"]`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Unauthorized" {
		t.Errorf("expected error message 'Unauthorized', got %s", response["error"])
	}
}

func TestDeleteURLs_InvalidUserID(t *testing.T) {
	mockRepo := mocks.NewLinkRepo(t)

	usecase := New(mockRepo, zap.NewNop(), DefaultConfig())
	defer usecase.Close()

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer([]byte(`["abc123"]`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Set("userID", 123)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Unauthorized" {
		t.Errorf("expected error message 'Unauthorized', got %s", response["error"])
	}
}

func TestDeleteURLs_RepositoryError(t *testing.T) {
	userID := "test-user-id"
	shortURLs := []string{"abc123", "def456"}

	mockRepo := mocks.NewLinkRepo(t)

	var wg sync.WaitGroup
	wg.Add(1)
	mockRepo.EXPECT().BatchDelete(mock.Anything, shortURLs, userID).Return(errors.New("database error")).Run(func(ctx context.Context, urls []string, uid string) {
		wg.Done()
	})

	usecase := New(mockRepo, zap.NewNop(), DefaultConfig())
	defer usecase.Close()

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer([]byte(`["abc123","def456"]`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	wg.Wait()
}

func TestDeleteURLs_BatchProcessing(t *testing.T) {
	userID := "test-user-id"

	shortURLs := make([]string, 1000)
	for i := range shortURLs {
		shortURLs[i] = "url" + string(rune(i))
	}

	mockRepo := mocks.NewLinkRepo(t)

	mockRepo.EXPECT().BatchDelete(mock.Anything, mock.AnythingOfType("[]string"), userID).Return(nil).Maybe()

	usecase := New(mockRepo, zap.NewNop(), DefaultConfig())
	defer usecase.Close()

	jsonData, _ := json.Marshal(shortURLs)
	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	time.Sleep(200 * time.Millisecond)
}
