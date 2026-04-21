package logging

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RequestLogger - middleware for logging http requests
func RequestLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t := time.Now()
			defer func() {
				logger.Info("http request",
					zap.String("uri", r.RequestURI),
					zap.String("method", r.Method),
					zap.Int("status", ww.Status()),
					zap.Duration("duration", time.Since(t)),
					zap.Int("size", ww.BytesWritten()))
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
