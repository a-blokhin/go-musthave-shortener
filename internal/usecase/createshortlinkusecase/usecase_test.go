package createshortlinkusecase

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/createshortlinkusecase/mocks"
)

func TestCreateShortURL_Success(t *testing.T) {
	baseURL := "http://example.com"
	testURL := "https://example.org/long/path"
	expectedAlias := "abcd1234"

	mockRepo := mocks.NewLinkRepo(t)
	mockRepo.On("Add", mock.Anything, testURL, "").Return(expectedAlias, nil)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	body := bytes.NewBufferString(testURL)
	req := httptest.NewRequest("POST", "/", body)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	expectedResponse, err := url.JoinPath(baseURL, expectedAlias)
	if err != nil {
		t.Errorf("Failed to create expected response: %v", err)
	}
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedResponse)
	}

	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/plain")
	}
}

func TestCreateShortURL_EmptyURL(t *testing.T) {
	baseURL := "http://example.com"
	mockRepo := mocks.NewLinkRepo(t)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	body := bytes.NewBufferString("")
	req := httptest.NewRequest("POST", "/", body)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	if !strings.Contains(rr.Body.String(), "URL is required") {
		t.Errorf("handler returned unexpected body: got %v, expected to contain 'URL is required'", rr.Body.String())
	}
}

func TestCreateShortURL_UsecaseError(t *testing.T) {
	baseURL := "http://example.com"
	testURL := "https://example.org/long/path"

	mockRepo := mocks.NewLinkRepo(t)
	mockRepo.On("Add", mock.Anything, testURL, "").Return("", io.EOF)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	body := bytes.NewBufferString(testURL)
	req := httptest.NewRequest("POST", "/", body)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}
