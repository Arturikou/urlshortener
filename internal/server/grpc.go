package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	pb "github.com/Arturikou/urlshortener/api/proto/shortener/v1"
	"github.com/Arturikou/urlshortener/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCServer struct {
	grpcServer *grpc.Server
	cfg        config.ServerConfig
	logger     *zap.Logger
}

func NewGRPC(
	cfg config.ServerConfig,
	service pb.ShortenerServiceServer,
	tlsConfig *tls.Config,
	logger *zap.Logger,
	interceptors ...grpc.UnaryServerInterceptor,
) *GRPCServer {
	opts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(interceptors...)}
	if tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterShortenerServiceServer(grpcServer, service)

	return &GRPCServer{
		grpcServer: grpcServer,
		cfg:        cfg,
		logger:     logger,
	}
}

func (s *GRPCServer) Run(ctx context.Context) error {
	listen, err := net.Listen("tcp", s.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	s.logger.Info("starting gRPC server", zap.String("addr", s.cfg.GRPCAddr))

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.grpcServer.Serve(listen)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.grpcServer.GracefulStop()
		return <-errCh
	}
}
