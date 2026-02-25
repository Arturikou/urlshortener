package handlers

import (
	"context"

	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/Arturikou/urlshortener/internal/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockery
type URLService interface {
	AddURL(ctx context.Context, originalURL string) (shortener.AddURLResult, error)
	GetURL(ctx context.Context, alias string) (string, error)
	AddURLs(ctx context.Context, shortenBatch []*models.ShortenBatch) ([]models.ShortenBatch, error)
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error)
}

type Handlers struct {
	urlService URLService
	pinger     storage.Pinger
	cfg        config.HandlersConfig
	logger     *zap.SugaredLogger
}

func New(urlService URLService, pinger storage.Pinger, cfg config.HandlersConfig, logger *zap.SugaredLogger) *Handlers {
	return &Handlers{
		urlService: urlService,
		pinger:     pinger,
		cfg:        cfg,
		logger:     logger,
	}
}
