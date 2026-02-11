package main

import (
	"context"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/logging"
	"github.com/Arturikou/urlshortener/internal/repository"
	"github.com/Arturikou/urlshortener/internal/repository/postgres"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	cfg := config.New()
	ctx := context.Background()

	l, err := logging.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("can't initialize logging: %v", err)
	}
	defer l.Sync()
	sl := l.Sugar()

	var repo *postgres.Storage
	if cfg.Database.DSN != "" {
		repo, err = postgres.New(ctx, cfg.Database.DSN)
		if err != nil {
			sl.Fatalf("can't initialize database: %v", err)
		}
		defer repo.Close()
	}

	memoryStorage := repository.NewMemoryStore()
	fileStorage, err := repository.NewFileStore(cfg.FileStoragePath, memoryStorage, sl)
	if err != nil {
		sl.Fatalf("failed to initialize file store: %v", err)
	}
	defer fileStorage.Close()

	if err := fileStorage.Load(); err != nil {
		sl.Warnf("could not restore data from file: %v", err)
	}

	urlService := shortener.New(fileStorage, sl)

	h := handlers.New(urlService, repo, cfg.Handlers, sl)
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
