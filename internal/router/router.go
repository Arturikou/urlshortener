package router

import (
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(h *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Post("/", h.AddURL)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetURL)
		})
	})

	return r
}
