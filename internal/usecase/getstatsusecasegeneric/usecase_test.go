package getstatsusecasegeneric

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/repository"
	"go-musthave-shortener/internal/usecase/getstatsusecasegeneric/mocks"
)

func TestGetStatsUsecase_Success(t *testing.T) {
	expectedStats := repository.Stats{
		URLs:  100,
		Users: 50,
	}

	mockRepo := &mocks.StatsRepository{}
	mockRepo.On("GetStats", mock.Anything).Return(expectedStats, nil)

	usecase := New(mockRepo, zap.NewNop())

	result, err := usecase.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, result)
	mockRepo.AssertExpectations(t)
}

func TestGetStatsUsecase_RepositoryError(t *testing.T) {
	mockRepo := &mocks.StatsRepository{}
	mockRepo.On("GetStats", mock.Anything).Return(repository.Stats{}, errors.New("repository error"))

	usecase := New(mockRepo, zap.NewNop())

	result, err := usecase.Execute(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "failed to get stats", err.Error())
	assert.Equal(t, repository.Stats{}, result)
	mockRepo.AssertExpectations(t)
}
