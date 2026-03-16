package grpcshortenurlusecase

import (
	"context"
	"errors"
	"net/url"
	"strings"

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
	baseURL  string
	audit    AuditEmitter
}

func New(linkRepo LinkRepo, logger *zap.Logger, baseURL string, audit AuditEmitter) *Usecase {
	return &Usecase{
		linkRepo: linkRepo,
		logger:   logger,
		baseURL:  baseURL,
		audit:    audit,
	}
}

func (u *Usecase) Execute(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		userID = ""
	}

	reqURL := strings.TrimSpace(req.GetUrl())
	if reqURL == "" {
		u.logger.Info("Empty URL provided in request")
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	alias, err := u.linkRepo.Add(ctx, reqURL, userID)
	if err != nil {
		var duplicateErr *model.DuplicateURLError
		if errors.As(err, &duplicateErr) {
			expResp, err := url.JoinPath(u.baseURL, duplicateErr.ExistingShortURL)
			if err != nil {
				u.logger.Error("Failed to create response", zap.Error(err))
				return nil, status.Error(codes.Internal, "failed to create short URL")
			}

			return &proto.URLShortenResponse{
				Result: expResp,
			}, nil
		}

		u.logger.Error("Failed to create short URL",
			zap.Error(err),
			zap.String("url", reqURL))
		return nil, status.Error(codes.Internal, "failed to create short URL")
	}

	expResp, err := url.JoinPath(u.baseURL, alias)
	if err != nil {
		u.logger.Error("Failed to create response", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create short URL")
	}

	if u.audit != nil {
		u.audit.Emit(audit.ActionShorten, userID, reqURL)
	}

	return &proto.URLShortenResponse{
		Result: expResp,
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
