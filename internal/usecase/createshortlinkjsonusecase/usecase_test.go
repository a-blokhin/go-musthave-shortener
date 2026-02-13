package createshortlinkjsonusecase_test

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

	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase"
	"go-musthave-shortener/internal/usecase/createshortlinkjsonusecase/mocks"
	"go-musthave-shortener/pkg/createshortlinkjsonpkg"
)

func TestUsecase_Execute(t *testing.T) {
	logger := zap.NewNop()
	baseURL := "http://localhost:8080"

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.LinkRepo)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "success",
			requestBody: createshortlinkjsonpkg.Request{
				URL: "https://practicum.yandex.ru",
			},
			mockSetup: func(m *mocks.LinkRepo) {
				m.On("Add", mock.Anything, "https://practicum.yandex.ru").Return("abc123", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: createshortlinkjsonpkg.Response{
				Result: "http://localhost:8080/abc123",
			},
		},
		{
			name: "empty URL",
			requestBody: createshortlinkjsonpkg.Request{
				URL: "",
			},
			mockSetup:      func(m *mocks.LinkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "URL is required"},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.LinkRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "Invalid JSON"},
		},
		{
			name: "repository error",
			requestBody: createshortlinkjsonpkg.Request{
				URL: "https://example.com",
			},
			mockSetup: func(m *mocks.LinkRepo) {
				m.On("Add", mock.Anything, "https://example.com").Return("", errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := mocks.NewLinkRepo(t)
			tt.mockSetup(mockRepo)

			usecase := createshortlinkjsonusecase.New(mockRepo, logger, baseURL)

			// Prepare request
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute
			usecase.Execute(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			var actualBody interface{}
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			assert.NoError(t, err)

			expectedBodyJSON, _ := json.Marshal(tt.expectedBody)
			var expectedBodyInterface interface{}
			json.Unmarshal(expectedBodyJSON, &expectedBodyInterface)

			assert.Equal(t, expectedBodyInterface, actualBody)

		})
	}
}
