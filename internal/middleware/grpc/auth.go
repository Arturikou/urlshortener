package grpc

import (
	"context"

	"github.com/Arturikou/urlshortener/internal/auth"
	"github.com/Arturikou/urlshortener/internal/userctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const authMetadataKey = "authorization"

type AuthInterceptor struct {
	logger *zap.SugaredLogger
}

func NewAuthInterceptor(logger *zap.SugaredLogger) *AuthInterceptor {
	return &AuthInterceptor{logger: logger}
}

func (a *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID, authenticated := a.userIDFromMetadata(ctx)
		if !authenticated {
			userID = uuid.New()
		}

		ctx = userctx.WithUserID(ctx, userID)
		ctx = userctx.WithAuthenticated(ctx, authenticated)

		return handler(ctx, req)
	}
}

func (a *AuthInterceptor) userIDFromMetadata(ctx context.Context) (uuid.UUID, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, false
	}

	values := md.Get(authMetadataKey)
	if len(values) == 0 || values[0] == "" {
		return uuid.Nil, false
	}

	userID, err := auth.GetUserID(values[0])
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}
