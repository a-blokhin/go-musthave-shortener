package createshortlinkjsonusecase

import "context"

//go:generate mockery --name=LinkRepo --output=./mocks --outpkg=mocks --filename=link_repo_mock.go --with-expecter
type LinkRepo interface {
	Add(ctx context.Context, url string, userID string) (alias string, err error)
}

//go:generate mockery --name=AuditEmitter --output=./mocks --outpkg=mocks --with-expecter --filename=audit_emitter_mock.go
type AuditEmitter interface {
	Emit(action, userID, url string)
}
