package createshortlinkjsonusecase

import "context"

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --with-expecter
type LinkRepo interface {
	Add(ctx context.Context, url string, userID string) (alias string, err error)
}
