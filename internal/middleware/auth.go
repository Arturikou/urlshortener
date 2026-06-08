package middleware

import (
	"net/http"

	"github.com/Arturikou/urlshortener/internal/auth"
	"github.com/Arturikou/urlshortener/internal/userctx"
	"github.com/google/uuid"
)

func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var userID uuid.UUID

		authCookie, err := r.Cookie("Authorization")
		if err == nil {
			userID, err = auth.GetUserID(authCookie.Value)
			if err == nil {
				next.ServeHTTP(w, r.WithContext(userctx.WithUserID(r.Context(), userID)))
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

		next.ServeHTTP(w, r.WithContext(userctx.WithUserID(r.Context(), userID)))
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

			next.ServeHTTP(w, r.WithContext(userctx.WithUserID(r.Context(), userID)))
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

		next.ServeHTTP(w, r.WithContext(userctx.WithUserID(r.Context(), userID)))
	})
}
