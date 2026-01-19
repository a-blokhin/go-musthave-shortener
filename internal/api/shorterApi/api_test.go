package shorterapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	createshortlinkusecase "go-musthave-shortener/internal/handler/createShortLinkUsecase"
	redirectfromshortlinkusecase "go-musthave-shortener/internal/handler/redirectFromShortLinkUsecase"
	createshortlinkpkg "go-musthave-shortener/pkg/createShortLinkPkg"
	redirectfromshortlinkpkg "go-musthave-shortener/pkg/redirectFromShortLinkPkg"
)

type MockLinkRepo struct {
	AddFunc func(url string) (string, error)
	GetFunc func(alias string) (string, error)
}

func (m *MockLinkRepo) Add(url string) (string, error) {
	if m.AddFunc != nil {
		return m.AddFunc(url)
	}
	return "", nil
}

func (m *MockLinkRepo) Get(alias string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(alias)
	}
	return "", nil
}

var _ createshortlinkusecase.LinkRepo = (*MockLinkRepo)(nil)
var _ redirectfromshortlinkusecase.LinkRepo = (*MockLinkRepo)(nil)

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

	api := New(
		baseURL,
		createshortlinkusecase.New(mockRepo),
		redirectfromshortlinkusecase.New(mockRepo),
	)

	body := bytes.NewBufferString(testURL)
	req := httptest.NewRequest(createshortlinkpkg.Method, createshortlinkpkg.MethodPath, body)
	rr := httptest.NewRecorder()

	api.handleRequest(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	expectedResponse := baseURL + "/" + expectedAlias
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedResponse)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/plain")
	}
}

func TestCreateShortURL_EmptyURL(t *testing.T) {

	baseURL := "http://example.com"
	mockRepo := &MockLinkRepo{}
	api := New(
		baseURL,
		createshortlinkusecase.New(mockRepo),
		redirectfromshortlinkusecase.New(mockRepo),
	)

	body := bytes.NewBufferString("")
	req := httptest.NewRequest(createshortlinkpkg.Method, createshortlinkpkg.MethodPath, body)
	rr := httptest.NewRecorder()

	api.handleRequest(rr, req)

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

	api := New(
		baseURL,
		createshortlinkusecase.New(mockRepo),
		redirectfromshortlinkusecase.New(mockRepo),
	)

	body := bytes.NewBufferString(testURL)
	req := httptest.NewRequest(createshortlinkpkg.Method, createshortlinkpkg.MethodPath, body)
	rr := httptest.NewRecorder()

	api.handleRequest(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestRedirectToOriginalURL_Success(t *testing.T) {

	baseURL := "http://example.com"
	testAlias := "abcd1234"
	originalURL := "https://example.org/long/path"

	mockRepo := &MockLinkRepo{
		GetFunc: func(alias string) (string, error) {
			if alias != testAlias {
				t.Errorf("Expected alias %s, got %s", testAlias, alias)
			}
			return originalURL, nil
		},
	}

	api := New(
		baseURL,
		createshortlinkusecase.New(mockRepo),
		redirectfromshortlinkusecase.New(mockRepo),
	)

	req := httptest.NewRequest(redirectfromshortlinkpkg.Method, "/"+testAlias, nil)
	rr := httptest.NewRecorder()

	api.handleRequest(rr, req)

	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	if location := rr.Header().Get("Location"); location != originalURL {
		t.Errorf("handler set wrong Location header: got %v want %v", location, originalURL)
	}
}

func TestRedirectToOriginalURL_EmptyAlias(t *testing.T) {

	baseURL := "http://example.com"
	mockRepo := &MockLinkRepo{}
	api := New(
		baseURL,
		createshortlinkusecase.New(mockRepo),
		redirectfromshortlinkusecase.New(mockRepo),
	)

	req := httptest.NewRequest(redirectfromshortlinkpkg.Method, "/", nil)
	rr := httptest.NewRecorder()

	api.handleRequest(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestRedirectToOriginalURL_UsecaseError(t *testing.T) {

	baseURL := "http://example.com"
	testAlias := "abcd1234"

	mockRepo := &MockLinkRepo{
		GetFunc: func(alias string) (string, error) {
			return "", io.EOF
		},
	}

	api := New(
		baseURL,
		createshortlinkusecase.New(mockRepo),
		redirectfromshortlinkusecase.New(mockRepo),
	)

	req := httptest.NewRequest(redirectfromshortlinkpkg.Method, "/"+testAlias, nil)
	rr := httptest.NewRecorder()

	api.handleRequest(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
