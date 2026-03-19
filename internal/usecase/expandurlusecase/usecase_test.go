package expandurlusecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/model"
	"go-musthave-shortener/internal/usecase/expandurlusecase/mocks"
)

func TestExpandURLUsecase_Success(t *testing.T) {
	alias := "abcd1234"
	expectedURL := "https://example.org/long/path"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("Get", mock.Anything, alias).Return(expectedURL, nil)

	usecase := New(mockRepo, zap.NewNop(), nil)

	result, err := usecase.Execute(context.Background(), alias, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, result)
	mockRepo.AssertExpectations(t)
}

func TestExpandURLUsecase_EmptyAlias(t *testing.T) {
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	usecase := New(mockRepo, zap.NewNop(), nil)

	result, err := usecase.Execute(context.Background(), "", userID)

	assert.Error(t, err)
	assert.Equal(t, "ID is required", err.Error())
	assert.Empty(t, result)
}

func TestExpandURLUsecase_URLNotFound(t *testing.T) {
	alias := "nonexistent"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	mockRepo.On("Get", mock.Anything, alias).Return("", errors.New("not found"))

	usecase := New(mockRepo, zap.NewNop(), nil)

	result, err := usecase.Execute(context.Background(), alias, userID)

	assert.Error(t, err)
	assert.Equal(t, "URL not found", err.Error())
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestExpandURLUsecase_URLDeleted(t *testing.T) {
	alias := "deleted123"
	userID := "user123"

	mockRepo := &mocks.LinkRepo{}
	deletedErr := &model.DeletedURLError{}
	mockRepo.On("Get", mock.Anything, alias).Return("", deletedErr)

	usecase := New(mockRepo, zap.NewNop(), nil)

	result, err := usecase.Execute(context.Background(), alias, userID)

	assert.Error(t, err)
	assert.Equal(t, "URL has been deleted", err.Error())
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
