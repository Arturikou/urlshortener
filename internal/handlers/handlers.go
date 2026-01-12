package handlers

import "github.com/Arturikou/urlshortener/internal/handlers/config"

//go:generate mockery
type URLService interface {
	AddURL(url string) (string, error)
	GetURL(id string) (string, error)
}

type Handlers struct {
	urlService URLService
	cfg        config.Config
}

func New(urlService URLService, cfg config.Config) *Handlers {
	return &Handlers{
		urlService: urlService,
		cfg:        cfg,
	}
}
