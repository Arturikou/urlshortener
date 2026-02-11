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
		r.Get("/ping", h.Ping)
		r.Post("/", h.AddURL)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetURL)
		})
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", h.Shorten)
		})
	})

	return r
}
