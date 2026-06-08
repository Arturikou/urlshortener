package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/Arturikou/urlshortener/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	cfg        config.ServerConfig
	logger     *zap.Logger
}

func New(cfg config.ServerConfig, handler http.Handler, tlsConfig *tls.Config, logger *zap.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:      cfg.Addr,
			Handler:   handler,
			TLSConfig: tlsConfig,
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
	if s.httpServer.TLSConfig != nil {
		s.logger.Info("starting HTTPS server", zap.String("addr", s.cfg.Addr))
		serve = func() error {
			// сертификат и ключ берутся из TLSConfig
			return s.httpServer.ListenAndServeTLS("", "")
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
