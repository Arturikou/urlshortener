package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	logger *zap.SugaredLogger
}

func NewLoggingInterceptor(logger *zap.SugaredLogger) *LoggingInterceptor {
	return &LoggingInterceptor{logger: logger}
}

func (l *LoggingInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		l.logger.Infow("gRPC request",
			"method", info.FullMethod,
			"code", status.Code(err).String(),
			"duration", time.Since(start),
		)

		return resp, err
	}
}
