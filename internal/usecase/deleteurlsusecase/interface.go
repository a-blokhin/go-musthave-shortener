package deleteurlsusecase

import "context"

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --with-expecter
type LinkRepo interface {
	BatchDelete(ctx context.Context, shortURLs []string, userID string) error
}
