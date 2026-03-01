package redirectfromshortlinkusecase

import "context"

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --filename=link_repo_mock.go --with-expecter
type LinkRepo interface {
	Get(ctx context.Context, alias string) (string, error)
}
