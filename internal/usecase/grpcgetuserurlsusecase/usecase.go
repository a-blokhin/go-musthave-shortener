package grpcgetuserurlsusecase

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go-musthave-shortener/api/proto"
	"go-musthave-shortener/internal/auth"
	"go-musthave-shortener/internal/usecase/getuserurlsusecasegeneric"
)

type Usecase struct {
	getUserURLsUsecase *getuserurlsusecasegeneric.GetUserURLsUsecase
	logger             *zap.Logger
}

func New(getUserURLsUsecase *getuserurlsusecasegeneric.GetUserURLsUsecase, logger *zap.Logger) *Usecase {
	return &Usecase{
		getUserURLsUsecase: getUserURLsUsecase,
		logger:             logger,
	}
}

func (u *Usecase) Execute(ctx context.Context) (*proto.UserURLsResponse, error) {
	userID, err := auth.GetUserIDFromGRPCContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	userURLs, err := u.getUserURLsUsecase.Execute(ctx, userID)
	if err != nil {
		u.logger.Error("Failed to get user URLs",
			zap.Error(err),
			zap.String("userID", userID))
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	u.logger.Info("Retrieved user URLs", zap.Int("count", len(userURLs)), zap.String("userID", userID))

	response := make([]*proto.URLData, len(userURLs))
	for i, userURL := range userURLs {
		response[i] = &proto.URLData{
			ShortUrl:    userURL.ShortURL,
			OriginalUrl: userURL.OriginalURL,
		}
	}

	return &proto.UserURLsResponse{
		Url: response,
	}, nil
}
