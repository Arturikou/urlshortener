package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/Arturikou/urlshortener/internal/auth"
	"github.com/google/uuid"
)

var ErrContextKeyNotFound = errors.New("key not found in context")
var ErrContextValueWrongType = errors.New("context value has unexpected type")

type ContextKey string

const ContextKeyUserID ContextKey = "userID"

func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var userID uuid.UUID

		authCookie, err := r.Cookie("Authorization")
		if err == nil {
			userID, err = auth.GetUserID(authCookie.Value)
			if err == nil {
				ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		userID = uuid.New()
		jwtToken, err := auth.BuildJWTString(userID)
		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "Authorization",
				Value: jwtToken,
				Path:  "/",
			})
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID uuid.UUID
		authCookie, err := r.Cookie("Authorization")
		if err == nil {
			userID, err = auth.GetUserID(authCookie.Value)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID = uuid.New()
		jwtToken, err := auth.BuildJWTString(userID)
		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "Authorization",
				Value: jwtToken,
				Path:  "/",
			})
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(ContextKeyUserID)
	if val == nil {
		return uuid.Nil, ErrContextKeyNotFound
	}

	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrContextValueWrongType
	}

	return id, nil
}
