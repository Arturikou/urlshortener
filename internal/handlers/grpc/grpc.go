package grpc

import (
	"context"

	pb "github.com/Arturikou/urlshortener/api/proto/shortener/v1"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockery
type UserURL interface {
	AddURL(ctx context.Context, originalURL string) (shortener.AddURLResult, error)
	GetURL(ctx context.Context, alias string) (string, error)
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error)
}

type Service struct {
	pb.UnimplementedShortenerServiceServer
	userURL UserURL
	cfg     config.HandlersConfig
	logger  *zap.SugaredLogger
}

func New(
	userURL UserURL,
	cfg config.HandlersConfig,
	logger *zap.SugaredLogger,
) *Service {
	return &Service{
		userURL: userURL,
		cfg:     cfg,
		logger:  logger,
	}
}
