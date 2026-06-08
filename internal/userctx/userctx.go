package userctx

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrKeyNotFound    = errors.New("key not found in context")
	ErrValueWrongType = errors.New("context value has unexpected type")
)

type contextKey string

const (
	keyUserID        contextKey = "userID"
	keyAuthenticated contextKey = "authenticated"
)

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, keyUserID, userID)
}

func WithAuthenticated(ctx context.Context, authenticated bool) context.Context {
	return context.WithValue(ctx, keyAuthenticated, authenticated)
}

func UserID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(keyUserID)
	if val == nil {
		return uuid.Nil, ErrKeyNotFound
	}

	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrValueWrongType
	}

	return id, nil
}

func IsAuthenticated(ctx context.Context) bool {
	authenticated, ok := ctx.Value(keyAuthenticated).(bool)
	return ok && authenticated
}
