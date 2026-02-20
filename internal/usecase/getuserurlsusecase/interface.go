package getuserurlsusecase

import (
	"context"

	"go-musthave-shortener/internal/repository"
)

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --filename=link_repo_mock.go --with-expecter
type LinkRepo interface {
	GetByUserID(ctx context.Context, userID string) ([]repository.UserURL, error)
}
