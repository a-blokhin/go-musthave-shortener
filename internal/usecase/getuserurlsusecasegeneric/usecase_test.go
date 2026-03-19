package getuserurlsusecasegeneric

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/repository"
	"go-musthave-shortener/internal/usecase/getuserurlsusecasegeneric/mocks"
)

func TestGetUserURLsUsecase_Success(t *testing.T) {
	baseURL := "http://example.com"
	userID := "user123"
	expectedURLs := []repository.UserURL{
		{ShortURL: "abc123", OriginalURL: "https://example.com/page1"},
		{ShortURL: "def456", OriginalURL: "https://example.com/page2"},
	}

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("GetByUserID", mock.Anything, userID).Return(expectedURLs, nil)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	result, err := usecase.Execute(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, baseURL+"/abc123", result[0].ShortURL)
	assert.Equal(t, "https://example.com/page1", result[0].OriginalURL)
	assert.Equal(t, baseURL+"/def456", result[1].ShortURL)
	assert.Equal(t, "https://example.com/page2", result[1].OriginalURL)
	mockRepo.AssertExpectations(t)
}

func TestGetUserURLsUsecase_NoURLs(t *testing.T) {
	baseURL := "http://example.com"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("GetByUserID", mock.Anything, userID).Return([]repository.UserURL{}, nil)

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	result, err := usecase.Execute(context.Background(), userID)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetUserURLsUsecase_RepositoryError(t *testing.T) {
	baseURL := "http://example.com"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("GetByUserID", mock.Anything, userID).Return([]repository.UserURL{}, errors.New("repository error"))

	usecase := New(mockRepo, zap.NewNop(), baseURL)

	result, err := usecase.Execute(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, "failed to get user URLs", err.Error())
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}
