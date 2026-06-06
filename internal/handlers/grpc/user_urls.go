package grpc

import (
	"context"
	"net/url"

	pb "github.com/Arturikou/urlshortener/api/proto/shortener/v1"
	"github.com/Arturikou/urlshortener/internal/userctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	if !userctx.IsAuthenticated(ctx) {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userID, err := userctx.UserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userURLs, err := s.userURL.GetUserURLs(ctx, userID)
	if err != nil {
		s.logger.Errorw("failed to get user URLs", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	if len(userURLs) == 0 {
		return &pb.UserURLsResponse{}, nil
	}

	items := make([]*pb.URLData, len(userURLs))
	for idx, userURL := range userURLs {
		shortURL, err := url.JoinPath(s.cfg.BaseAddr, userURL.Alias)
		if err != nil {
			s.logger.Errorw("failed to build short URL path", "base", s.cfg.BaseAddr, "alias", userURL.Alias, "error", err)
			return nil, status.Error(codes.Internal, "internal error")
		}

		item := &pb.URLData{}
		item.SetShortUrl(shortURL)
		item.SetOriginalUrl(userURL.OriginalURL)
		items[idx] = item
	}

	resp := &pb.UserURLsResponse{}
	resp.SetUrl(items)

	return resp, nil
}
