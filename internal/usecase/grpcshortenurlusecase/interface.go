package grpcshortenurlusecase

import (
	"context"
)

type LinkRepo interface {
	Add(ctx context.Context, url string, userID string) (string, error)
}

type AuditEmitter interface {
	Emit(action, userID, url string)
}
