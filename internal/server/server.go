package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/crypto"
	"github.com/Arturikou/urlshortener/internal/utils"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	cfg        config.ServerConfig
	logger     *zap.Logger
}

func New(cfg config.ServerConfig, handler http.Handler, logger *zap.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		},
		cfg:    cfg,
		logger: logger,
	}
}

// RunPprof запускает pprof-сервер
func (s *Server) RunPprof(ctx context.Context) error {
	srv := &http.Server{Addr: s.cfg.PprofAddr}
	s.logger.Info("starting pprof server", zap.String("addr", s.cfg.PprofAddr))
	return runWithShutdown(ctx, srv, s.cfg.ShutdownTimeout, srv.ListenAndServe)
}

// Run запускает HTTP(S) сервер
func (s *Server) Run(ctx context.Context) error {
	serve := s.httpServer.ListenAndServe
	if s.cfg.EnableHTTPS {
		if !utils.FileExists(s.cfg.CertFile) || !utils.FileExists(s.cfg.KeyFile) {
			s.logger.Info("TLS certificates not found, will be generated")
			if err := crypto.Generate(s.cfg.CertFile, s.cfg.KeyFile); err != nil {
				return fmt.Errorf("tls setup failed: %w", err)
			}
		}

		s.logger.Info("starting HTTPS server",
			zap.String("addr", s.cfg.Addr),
			zap.String("cert", s.cfg.CertFile),
		)

		serve = func() error {
			return s.httpServer.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile)
		}
	} else {
		s.logger.Info("starting HTTP server", zap.String("addr", s.cfg.Addr))
	}

	return runWithShutdown(ctx, s.httpServer, s.cfg.ShutdownTimeout, serve)
}

// runWithShutdown завершает сервер при отмене контекста
func runWithShutdown(
	ctx context.Context,
	srv *http.Server,
	timeout time.Duration,
	serve func() error,
) error {
	errCh := make(chan error, 1)
	go func() {
		if err := serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return <-errCh
	}
}
