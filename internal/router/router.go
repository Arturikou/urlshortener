package router

import (
	"github.com/Arturikou/urlshortener/internal/handlers"
	"net/http"
)

func New(h *handlers.Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.AddURL)
	mux.HandleFunc("GET /{id}", h.GetURL)
	return mux
}
