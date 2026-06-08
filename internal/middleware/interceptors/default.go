package interceptors

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func DefaultUnaryInterceptors(logger *zap.SugaredLogger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		NewLoggingInterceptor(logger).UnaryServerInterceptor(),
		NewAuthInterceptor(logger).UnaryServerInterceptor(),
	}
}
