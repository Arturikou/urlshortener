package middleware

import (
	"context"
	"net/http"

	"github.com/Arturikou/urlshortener/internal/auth"
	"github.com/google/uuid"
)

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

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return id, ok
}
