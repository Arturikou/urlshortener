package main

import (
	"context"
	"log"
	"net/http"

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
	auditManager := audit.NewManager()
	if cfg.Audit.AuditFile != "" {
		fileObs, err := audit.NewFileObserver(cfg.Audit.AuditFile, sl)
		if err != nil {
			log.Fatalf("failed to init file auditor: %v", err)
		}

		auditManager.Register(fileObs)
	}
	if cfg.Audit.AuditURL != "" {
		client := httpaudit.New(cfg.Audit.AuditURL, sl)
		httpObs := audit.NewHTTPObserver(client, sl)
		auditManager.Register(httpObs)
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

	l.Info("starting server", zap.String("addr", cfg.ServerAddr))
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
