package grpcapi

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"go-musthave-shortener/api/proto"
	"go-musthave-shortener/internal/usecase/grpcexpandurlusecase"
	"go-musthave-shortener/internal/usecase/grpcgetuserurlsusecase"
	"go-musthave-shortener/internal/usecase/grpcshortenurlusecase"
)

type Server struct {
	proto.UnimplementedShortenerServiceServer
	shortenURLUseCase  *grpcshortenurlusecase.Usecase
	expandURLUseCase   *grpcexpandurlusecase.Usecase
	getUserURLsUseCase *grpcgetuserurlsusecase.Usecase
}

func New(
	shortenURLUseCase *grpcshortenurlusecase.Usecase,
	expandURLUseCase *grpcexpandurlusecase.Usecase,
	getUserURLsUseCase *grpcgetuserurlsusecase.Usecase,
) *Server {
	return &Server{
		shortenURLUseCase:  shortenURLUseCase,
		expandURLUseCase:   expandURLUseCase,
		getUserURLsUseCase: getUserURLsUseCase,
	}
}

func (s *Server) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	return s.shortenURLUseCase.Execute(ctx, req)
}

func (s *Server) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	return s.expandURLUseCase.Execute(ctx, req)
}

func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	return s.getUserURLsUseCase.Execute(ctx)
}
