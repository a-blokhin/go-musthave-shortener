package grpcgetuserurlsusecase

import (
	"context"
	"errors"
	"net/url"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go-musthave-shortener/api/proto"
)

type Usecase struct {
	linkRepo LinkRepo
	logger   *zap.Logger
	baseURL  string
}

func New(linkRepo LinkRepo, logger *zap.Logger, baseURL string) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
		logger:   logger,
		baseURL:  baseURL,
	}
}

func (u *Usecase) Execute(ctx context.Context) (*proto.UserURLsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	userURLs, err := u.linkRepo.GetByUserID(ctx, userID)
	if err != nil {
		u.logger.Error("Failed to get user URLs",
			zap.Error(err),
			zap.String("userID", userID))
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	u.logger.Info("Retrieved user URLs", zap.Int("count", len(userURLs)), zap.String("userID", userID))

	response := make([]*proto.URLData, len(userURLs))
	for i, userURL := range userURLs {
		expectedResponse, err := url.JoinPath(u.baseURL, userURL.ShortURL)
		if err != nil {
			u.logger.Error("Failed to create response", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to create response")
		}
		response[i] = &proto.URLData{
			ShortUrl:    expectedResponse,
			OriginalUrl: userURL.OriginalURL,
		}
	}

	return &proto.UserURLsResponse{
		Url: response,
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
