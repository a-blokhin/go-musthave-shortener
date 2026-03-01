package getuserurlsusecase

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/repository"
	"go-musthave-shortener/internal/usecase/getuserurlsusecase/mocks"
)

func TestGetUserURLs_Success(t *testing.T) {
	baseURL := "http://example.com"
	userID := "user123"
	userURLs := []repository.UserURL{
		{ShortURL: "abc123", OriginalURL: "https://example1.com"},
		{ShortURL: "def456", OriginalURL: "https://example2.com"},
	}

	mockRepo := mocks.NewLinkRepo(t)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(userURLs, nil)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response []repository.UserURL
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(response) != len(userURLs) {
		t.Errorf("expected %d URLs in response, got %d", len(userURLs), len(response))
	}

	for i, userURL := range response {
		expectedShortURL, err := url.JoinPath(baseURL, userURLs[i].ShortURL)
		if err != nil {
			t.Errorf("Failed to create expected short URL: %v", err)
		}
		if userURL.ShortURL != expectedShortURL {
			t.Errorf("expected short URL %s, got %s", expectedShortURL, userURL.ShortURL)
		}
		if userURL.OriginalURL != userURLs[i].OriginalURL {
			t.Errorf("expected original URL %s, got %s", userURLs[i].OriginalURL, userURL.OriginalURL)
		}
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}

func TestGetUserURLs_NoURLs(t *testing.T) {
	baseURL := "http://example.com"
	userID := "user123"
	userURLs := []repository.UserURL{}

	mockRepo := mocks.NewLinkRepo(t)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(userURLs, nil)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	t.Logf("Response status: %d", rr.Code)
	t.Logf("Response body: %s", rr.Body.String())

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestGetUserURLs_NoUserID(t *testing.T) {
	baseURL := "http://example.com"

	mockRepo := mocks.NewLinkRepo(t)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
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

func TestGetUserURLs_RepositoryError(t *testing.T) {
	baseURL := "http://example.com"
	userID := "user123"

	mockRepo := mocks.NewLinkRepo(t)
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, errors.New("database error"))

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	ctx.Set("userID", userID)

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Internal Server Error" {
		t.Errorf("expected error message 'Internal Server Error', got %s", response["error"])
	}
}
