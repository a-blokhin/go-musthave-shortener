package grpcexpandurlusecase

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go-musthave-shortener/api/proto"
	"go-musthave-shortener/internal/auth"
	"go-musthave-shortener/internal/usecase/expandurlusecase"
)

type Usecase struct {
	expandURLUsecase *expandurlusecase.ExpandURLUsecase
	logger           *zap.Logger
}

func New(expandURLUsecase *expandurlusecase.ExpandURLUsecase, logger *zap.Logger) *Usecase {
	return &Usecase{
		expandURLUsecase: expandURLUsecase,
		logger:           logger,
	}
}

func (u *Usecase) Execute(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	userID, err := auth.GetUserIDFromGRPCContext(ctx)
	if err != nil {
		userID = ""
	}

	alias := req.GetId()
	if alias == "" {
		u.logger.Info("Empty alias provided in request")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	result, err := u.expandURLUsecase.Execute(ctx, alias, userID)
	if err != nil {
		u.logger.Error("Failed to expand URL",
			zap.Error(err),
			zap.String("alias", alias))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &proto.URLExpandResponse{
		Result: result,
	}, nil
}
