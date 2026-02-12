package createshortlinkbatchusecase_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/usecase/createshortlinkbatchusecase"
	"go-musthave-shortener/pkg/createshortlinkbatchpkg"
)

type MockLinkRepo struct {
	mock.Mock
}

func (m *MockLinkRepo) AddBatch(urls []string) ([]string, error) {
	args := m.Called(urls)
	return args.Get(0).([]string), args.Error(1)
}

func TestUsecase_Execute(t *testing.T) {
	logger := zap.NewNop()
	baseURL := "http://localhost:8080"

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockLinkRepo)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "success - single item",
			requestBody: createshortlinkbatchpkg.BatchRequest{
				{
					CorrelationID: "id1",
					OriginalURL:   "https://practicum.yandex.ru",
				},
			},
			mockSetup: func(m *MockLinkRepo) {
				m.On("AddBatch", []string{"https://practicum.yandex.ru"}).Return([]string{"abc123"}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: createshortlinkbatchpkg.BatchResponse{
				{
					CorrelationID: "id1",
					ShortURL:      "http://localhost:8080/abc123",
				},
			},
		},
		{
			name: "success - multiple items",
			requestBody: createshortlinkbatchpkg.BatchRequest{
				{
					CorrelationID: "id1",
					OriginalURL:   "https://practicum.yandex.ru",
				},
				{
					CorrelationID: "id2",
					OriginalURL:   "https://example.com",
				},
			},
			mockSetup: func(m *MockLinkRepo) {
				m.On("AddBatch", []string{"https://practicum.yandex.ru", "https://example.com"}).Return([]string{"abc123", "def456"}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: createshortlinkbatchpkg.BatchResponse{
				{
					CorrelationID: "id1",
					ShortURL:      "http://localhost:8080/abc123",
				},
				{
					CorrelationID: "id2",
					ShortURL:      "http://localhost:8080/def456",
				},
			},
		},
		{
			name:           "empty batch",
			requestBody:    createshortlinkbatchpkg.BatchRequest{},
			mockSetup:      func(m *MockLinkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "Batch cannot be empty"},
		},
		{
			name: "empty URL in batch item",
			requestBody: createshortlinkbatchpkg.BatchRequest{
				{
					CorrelationID: "id1",
					OriginalURL:   "",
				},
			},
			mockSetup:      func(m *MockLinkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "URL cannot be empty"},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockLinkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "Invalid JSON"},
		},
		{
			name: "repository error",
			requestBody: createshortlinkbatchpkg.BatchRequest{
				{
					CorrelationID: "id1",
					OriginalURL:   "https://example.com",
				},
			},
			mockSetup: func(m *MockLinkRepo) {
				m.On("AddBatch", []string{"https://example.com"}).Return([]string{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := new(MockLinkRepo)
			tt.mockSetup(mockRepo)

			usecase := createshortlinkbatchusecase.New(mockRepo, logger, baseURL)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			usecase.Execute(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			var actualBody interface{}
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			assert.NoError(t, err)

			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			var expectedBodyInterface interface{}
			json.Unmarshal(expectedBodyJSON, &expectedBodyInterface)

			assert.Equal(t, expectedBodyInterface, actualBody)

			mockRepo.AssertExpectations(t)
		})
	}
}
