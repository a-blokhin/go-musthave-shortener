package shortenurlusecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/model"
	"go-musthave-shortener/internal/usecase/shortenurlusecase/mocks"
)

func TestShortenURLUsecase_Success(t *testing.T) {
	baseURL := "http://example.com"
	testURL := "https://example.org/long/path"
	expectedAlias := "abcd1234"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("Add", mock.Anything, testURL, userID).Return(expectedAlias, nil)

	usecase := New(mockRepo, zap.NewNop(), baseURL, nil)

	result, err := usecase.Execute(context.Background(), testURL, userID)

	assert.NoError(t, err)
	expectedResult := baseURL + "/" + expectedAlias
	assert.Equal(t, expectedResult, result)
	mockRepo.AssertExpectations(t)
}

func TestShortenURLUsecase_DuplicateURL(t *testing.T) {
	baseURL := "http://example.com"
	testURL := "https://example.org/long/path"
	existingAlias := "existing123"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	duplicateErr := &model.DuplicateURLError{ExistingShortURL: existingAlias}
	mockRepo.On("Add", mock.Anything, testURL, userID).Return("", duplicateErr)

	usecase := New(mockRepo, zap.NewNop(), baseURL, nil)

	result, err := usecase.Execute(context.Background(), testURL, userID)

	assert.NoError(t, err)
	expectedResult := baseURL + "/" + existingAlias
	assert.Equal(t, expectedResult, result)
	mockRepo.AssertExpectations(t)
}

func TestShortenURLUsecase_EmptyURL(t *testing.T) {
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	usecase := New(mockRepo, zap.NewNop(), "http://example.com", nil)

	result, err := usecase.Execute(context.Background(), "", userID)

	assert.Error(t, err)
	assert.Equal(t, "URL is required", err.Error())
	assert.Empty(t, result)
}

func TestShortenURLUsecase_RepositoryError(t *testing.T) {
	baseURL := "http://example.com"
	testURL := "https://example.org/long/path"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("Add", mock.Anything, testURL, userID).Return("", errors.New("repository error"))

	usecase := New(mockRepo, zap.NewNop(), baseURL, nil)

	result, err := usecase.Execute(context.Background(), testURL, userID)

	assert.Error(t, err)
	assert.Equal(t, "failed to create short URL", err.Error())
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestShortenURLUsecase_BaseURLJoinError(t *testing.T) {
	testURL := "https://example.org/long/path"
	expectedAlias := "abcd1234"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("Add", mock.Anything, testURL, userID).Return(expectedAlias, nil)

	// Test with invalid base URL that would cause url.JoinPath to fail
	// Using a base URL with invalid port format that would cause url.JoinPath to fail
	usecase := New(mockRepo, zap.NewNop(), "http://example.com:invalid:port", nil)

	result, err := usecase.Execute(context.Background(), testURL, userID)

	assert.Error(t, err)
	assert.Equal(t, "failed to create short URL", err.Error())
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
