package grpcshortenurlusecase

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go-musthave-shortener/api/proto"
	"go-musthave-shortener/internal/auth"
	"go-musthave-shortener/internal/usecase/shortenurlusecase"
)

type Usecase struct {
	shortenURLUsecase *shortenurlusecase.ShortenURLUsecase
	logger            *zap.Logger
}

func New(shortenURLUsecase *shortenurlusecase.ShortenURLUsecase, logger *zap.Logger) *Usecase {
	return &Usecase{
		shortenURLUsecase: shortenURLUsecase,
		logger:            logger,
	}
}

func (u *Usecase) Execute(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	userID, err := auth.GetUserIDFromGRPCContext(ctx)
	if err != nil {
		userID = ""
	}

	reqURL := req.GetUrl()
	if reqURL == "" {
		u.logger.Info("Empty URL provided in request")
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	result, err := u.shortenURLUsecase.Execute(ctx, reqURL, userID)
	if err != nil {
		u.logger.Error("Failed to create short URL",
			zap.Error(err),
			zap.String("url", reqURL))
		return nil, status.Error(codes.Internal, "failed to create short URL")
	}

	return &proto.URLShortenResponse{
		Result: result,
	}, nil
}
