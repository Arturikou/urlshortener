package handlers

import (
	"context"
	"github.com/Arturikou/urlshortener/internal/config"
	"go.uber.org/zap"
)

//go:generate mockery
type URLService interface {
	AddURL(url string) (string, error)
	GetURL(id string) (string, error)
}

//go:generate mockery
type Repo interface {
	Ping(ctx context.Context) error
}

type Handlers struct {
	urlService URLService
	repo       Repo
	cfg        config.HandlersConfig
	logger     *zap.SugaredLogger
}

func New(urlService URLService, repo Repo, cfg config.HandlersConfig, logger *zap.SugaredLogger) *Handlers {
	return &Handlers{
		urlService: urlService,
		repo:       repo,
		cfg:        cfg,
		logger:     logger,
	}
}
