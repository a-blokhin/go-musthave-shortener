package createshortlinkbatchusecase

import "context"

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --with-expecter
type LinkRepo interface {
	AddBatch(ctx context.Context, urls []string) ([]string, error)
}
