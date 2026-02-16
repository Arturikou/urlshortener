package handlers

import (
	"context"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/Arturikou/urlshortener/internal/storage"
	"go.uber.org/zap"
)

//go:generate mockery
type URLService interface {
	AddURL(ctx context.Context, url string) (string, error)
	GetURL(ctx context.Context, alias string) (string, error)
	AddURLs(ctx context.Context, shortenBatch []*shortener.ShortenBatch) ([]shortener.ShortenBatch, error)
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
