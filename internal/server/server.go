package server

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

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

func (s *Server) RunPprof() {
	s.logger.Info("starting pprof server", zap.String("addr", s.cfg.PprofAddr))
	if err := http.ListenAndServe(s.cfg.PprofAddr, nil); err != nil {
		s.logger.Warn("pprof server error", zap.Error(err))
	}
}

func (s *Server) Run() error {
	if !s.cfg.EnableHTTPS {
		s.logger.Info("starting HTTP server", zap.String("addr", s.cfg.Addr))
		return s.httpServer.ListenAndServe()
	}

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

	return s.httpServer.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile)
}
