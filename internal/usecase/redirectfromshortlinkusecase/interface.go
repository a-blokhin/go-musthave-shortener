package redirectfromshortlinkusecase

import "context"

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --with-expecter
type LinkRepo interface {
	Get(ctx context.Context, alias string) (string, error)
}
