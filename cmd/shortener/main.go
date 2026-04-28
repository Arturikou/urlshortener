// Package main URL Shortener API.
//
//	@title					URL Shortener API
//	@version				1.0
//	@description			A URL shortening service with JWT cookie authentication.
package main

import (
	"context"
	"log"
	"net/http"

	_ "github.com/Arturikou/urlshortener/docs"
	"github.com/Arturikou/urlshortener/internal/clients/httpaudit"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/logging"
	"github.com/Arturikou/urlshortener/internal/managers/audit"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/Arturikou/urlshortener/internal/storage"
	"github.com/Arturikou/urlshortener/internal/workers/deleteworker"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("can't load config: %v", err)
	}
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
	auditManager := audit.NewManager(ctx)
	if cfg.Audit.AuditFile != "" {
		fileObs, err := audit.NewFileObserver(cfg.Audit.AuditFile, sl)
		if err != nil {
			log.Fatalf("failed to init file auditor: %v", err)
		}

		auditManager.Register(fileObs)
	}
	if cfg.Audit.AuditURL != "" {
		client := httpaudit.New(cfg.Audit.AuditURL, sl)
		auditManager.Register(client)
	}

	urlService := shortener.New(st.URLRepo, st.UserURLRepo, st.Transactor, auditManager, sl)
	worker := deleteworker.New(urlService, sl)
	go worker.Run(ctx)

	h := handlers.New(
		urlService,
		worker,
		st.Pinger,
		cfg.Handlers,
		sl,
	)
	r := router.New(h, l)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	go func() {
		log.Println("pprof listening on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			sl.Warnf("pprof server error: %v", err)
		}
	}()

	l.Info("starting server", zap.String("addr", cfg.ServerAddr))
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
