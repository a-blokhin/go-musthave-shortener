package grpcexpandurlusecase

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go-musthave-shortener/api/proto"
	"go-musthave-shortener/internal/audit"
	"go-musthave-shortener/internal/model"
)

type Usecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
	audit    AuditEmitter
}

func New(linkRepo LinkRepo, logger *zap.Logger, audit AuditEmitter) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
		logger:   logger,
		audit:    audit,
	}
}

func (u *Usecase) Execute(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		userID = ""
	}

	alias := req.GetId()
	if alias == "" {
		u.logger.Info("Empty alias provided in request")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	originalURL, err := u.linkRepo.Get(ctx, alias)
	if err != nil {
		u.logger.Info("URL not found for alias",
			zap.String("alias", alias),
			zap.Error(err))

		var deletedErr *model.DeletedURLError
		if errors.As(err, &deletedErr) {
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		}

		return nil, status.Error(codes.NotFound, "URL not found")
	}

	if u.audit != nil {
		u.audit.Emit(audit.ActionFollow, userID, originalURL)
	}

	return &proto.URLExpandResponse{
		Result: originalURL,
	}, nil
}

func getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata in context")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", errors.New("no authorization header")
	}

	authHeader := authHeaders[0]
	if authHeader == "" {
		return "", errors.New("empty authorization header")
	}

	return authHeader, nil
}
