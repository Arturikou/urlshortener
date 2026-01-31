package main

import (
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/logging"
	"github.com/Arturikou/urlshortener/internal/repository"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	cfg := config.New()

	l, err := logging.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("can't initialize logging: %v", err)
	}
	defer l.Sync()
	sl := l.Sugar()

	storage, err := repository.NewStore(cfg.FileStoragePath)
	if err != nil {
		log.Fatalf("can't initialize storage: %v", err)
	}

	urlService := shortener.New(storage, sl)
	h := handlers.New(urlService, cfg.Handlers, sl)
	r := router.New(h, l)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	l.Info("starting server", zap.String("addr", cfg.ServerAddr))
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
