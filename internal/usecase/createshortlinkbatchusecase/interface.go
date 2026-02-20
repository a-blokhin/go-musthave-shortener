package createshortlinkbatchusecase

import "context"

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --filename=link_repo_mock.go --with-expecter
type LinkRepo interface {
	AddBatch(ctx context.Context, urls []string, userID string) ([]string, error)
}
