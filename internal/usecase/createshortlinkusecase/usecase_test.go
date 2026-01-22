package createshortlinkusecase

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MockLinkRepo struct {
	AddFunc func(url string) (string, error)
}

func (m *MockLinkRepo) Add(url string) (string, error) {
	if m.AddFunc != nil {
		return m.AddFunc(url)
	}
	return "", nil
}

func TestCreateShortURL_Success(t *testing.T) {
	baseURL := "http://example.com"
	testURL := "https://example.org/long/path"
	expectedAlias := "abcd1234"

	mockRepo := &MockLinkRepo{
		AddFunc: func(url string) (string, error) {
			if url != testURL {
				t.Errorf("Expected URL %s, got %s", testURL, url)
			}
			return expectedAlias, nil
		},
	}

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

	expectedResponse := baseURL + "/" + expectedAlias
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedResponse)
	}

	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/plain")
	}
}

func TestCreateShortURL_EmptyURL(t *testing.T) {
	baseURL := "http://example.com"
	mockRepo := &MockLinkRepo{}

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

	mockRepo := &MockLinkRepo{
		AddFunc: func(url string) (string, error) {
			return "", io.EOF
		},
	}

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
