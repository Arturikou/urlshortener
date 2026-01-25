package router

import (
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func New(h *handlers.Handlers, l *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger.WithLogging(l))

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.AddURL)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetURL)
		})
	})

	return r
}
