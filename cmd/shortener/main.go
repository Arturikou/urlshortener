package main

import (
	"context"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/logging"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/Arturikou/urlshortener/internal/storage"
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

	st, err := storage.New(ctx, cfg, sl)
	if err != nil {
		sl.Fatalf("failed to init st: %v", err)
	}

	if st.Closer != nil {
		defer st.Closer()
	}

	urlService := shortener.New(st.URLRepo, sl)

	h := handlers.New(urlService, st.Pinger, cfg.Handlers, sl)
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
