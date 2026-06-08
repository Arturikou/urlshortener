// Package main URL Shortener API.
//
//	@title					URL Shortener API
//	@version				1.0
//	@description			A URL shortening service with JWT cookie authentication.
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	_ "github.com/Arturikou/urlshortener/docs"
	"github.com/Arturikou/urlshortener/internal/clients/httpaudit"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/crypto"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/handlers/grpchandler"
	"github.com/Arturikou/urlshortener/internal/logging"
	"github.com/Arturikou/urlshortener/internal/managers/audit"
	"github.com/Arturikou/urlshortener/internal/middleware/interceptors"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/server"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/Arturikou/urlshortener/internal/storage"
	"github.com/Arturikou/urlshortener/internal/workers/deleteworker"
	"golang.org/x/sync/errgroup"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("can't load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

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
	grpcURLService := grpchandler.New(urlService, cfg.Handlers, sl)

	h := handlers.New(
		urlService,
		worker,
		st.Pinger,
		cfg.Handlers,
		sl,
	)
	r := router.New(h, l, cfg.TrustedSubnet)

	var tlsConfig *tls.Config
	if cfg.Server.EnableHTTPS {
		tlsConfig, err = crypto.LoadTLSConfig(cfg.Server.CertFile, cfg.Server.KeyFile)
		if err != nil {
			log.Fatalf("tls setup failed: %v", err)
		}
	}

	httpSrv := server.New(cfg.Server, r, tlsConfig, l)

	grpcSrv := server.NewGRPC(cfg.Server, grpcURLService, tlsConfig, l,
		interceptors.DefaultUnaryInterceptors(sl)...,
	)

	// Завершение по сигналу
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		worker.Run(gCtx)

		return nil
	})

	g.Go(func() error {
		if err := httpSrv.RunPprof(gCtx); err != nil {
			sl.Warnf("pprof server: %v", err)
		}

		return nil
	})

	g.Go(func() error {
		return httpSrv.Run(gCtx)
	})

	g.Go(func() error {
		return grpcSrv.Run(gCtx)
	})

	if err := g.Wait(); err != nil {
		sl.Errorf("server fatal error: %v", err)
	}
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
