package grpcexpandurlusecase

import (
	"context"
)

type LinkRepo interface {
	Get(ctx context.Context, alias string) (string, error)
}

type AuditEmitter interface {
	Emit(action, userID, url string)
}
