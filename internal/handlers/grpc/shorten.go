package grpc

import (
	"context"
	"errors"
	"net/url"

	pb "github.com/Arturikou/urlshortener/api/proto/shortener/v1"
	"github.com/Arturikou/urlshortener/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	originalURL := req.GetUrl()

	if originalURL == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	addResult, err := s.userURL.AddURL(ctx, originalURL)
	if err != nil {
		s.logger.Errorw("error add shortening URL", "url", originalURL, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	shortURL, err := url.JoinPath(s.cfg.BaseAddr, addResult.Alias)
	if err != nil {
		s.logger.Errorw("failed to build short URL path", "base", s.cfg.BaseAddr, "alias", addResult.Alias, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &pb.URLShortenResponse{}
	resp.SetResult(shortURL)

	return resp, nil
}

func (s *Service) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	alias := req.GetId()
	if alias == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	originalURL, err := s.userURL.GetURL(ctx, alias)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "shortener not found")
		}

		if errors.Is(err, models.ErrURLDeleted) {
			return nil, status.Error(codes.NotFound, "shortener has been deleted")
		}

		s.logger.Errorw("database error during ExpandURL", "alias", alias, "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &pb.URLExpandResponse{}
	resp.SetResult(originalURL)

	return resp, nil
}
