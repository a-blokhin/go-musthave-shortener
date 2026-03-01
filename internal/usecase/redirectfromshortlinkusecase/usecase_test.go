package redirectfromshortlinkusecase

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase/mocks"
)

func TestRedirectToOriginalURL_Success(t *testing.T) {
	testAlias := "abcd1234"
	originalURL := "https://example.org/long/path"

	mockRepo := mocks.NewLinkRepo(t)
	mockRepo.On("Get", mock.Anything, testAlias).Return(originalURL, nil)

	usecase := New(mockRepo, zap.NewNop(), nil)

	req := httptest.NewRequest("GET", "/"+testAlias, nil)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Params = gin.Params{gin.Param{Key: "id", Value: testAlias}}

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	if location := rr.Header().Get("Location"); location != originalURL {
		t.Errorf("handler set wrong Location header: got %v want %v", location, originalURL)
	}
}

func TestRedirectToOriginalURL_EmptyAlias(t *testing.T) {
	mockRepo := mocks.NewLinkRepo(t)

	usecase := New(mockRepo, zap.NewNop(), nil)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Params = gin.Params{gin.Param{Key: "id", Value: ""}}

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestRedirectToOriginalURL_UsecaseError(t *testing.T) {
	testAlias := "abcd1234"

	mockRepo := mocks.NewLinkRepo(t)
	mockRepo.On("Get", mock.Anything, testAlias).Return("", http.ErrMissingFile)

	usecase := New(mockRepo, zap.NewNop(), nil)

	req := httptest.NewRequest("GET", "/"+testAlias, nil)
	rr := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(rr)
	ctx.Request = req
	ctx.Params = gin.Params{gin.Param{Key: "id", Value: testAlias}}

	usecase.Execute(ctx)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}
