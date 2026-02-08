package handlers

import (
	"github.com/Arturikou/urlshortener/internal/config"
	"go.uber.org/zap"
)

//go:generate mockery
type URLService interface {
	AddURL(url string) (string, error)
	GetURL(id string) (string, error)
}

type Handlers struct {
	urlService URLService
	cfg        config.HandlersConfig
	logger     *zap.SugaredLogger
}

func New(
	urlService URLService,
	cfg config.HandlersConfig,
	logger *zap.SugaredLogger,
) *Handlers {
	return &Handlers{
		urlService: urlService,
		cfg:        cfg,
		logger:     logger,
	}
}
