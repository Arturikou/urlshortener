package router

import (
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/logging"
	mw "github.com/Arturikou/urlshortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func New(h *handlers.Handlers, l *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logging.RequestLogger(l))
	r.Use(mw.Gzip)

	r.Route("/", func(r chi.Router) {
		r.With(mw.OptionalAuth).Get("/ping", h.Ping)
		r.With(mw.OptionalAuth).Post("/", h.AddURL)
		r.With(mw.OptionalAuth).Get("/{alias}", h.GetURL)

		r.Route("/api", func(r chi.Router) {
			r.With(mw.OptionalAuth).Post("/shorten", h.Shorten)
			r.With(mw.OptionalAuth).Post("/shorten/batch", h.ShortenBatch)
			r.With(mw.RequireAuth).Get("/user/urls", h.UserUrls)
			r.With(mw.RequireAuth).Delete("/user/urls", h.DeleteUserURLs)
		})
	})

	return r
}
