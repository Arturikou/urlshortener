package router

import (
	"github.com/Arturikou/urlshortener/internal/handler"
	"net/http"
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, rootHandler)
	return mux
}

// Временное решение пока не перейду на другой роутер
func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.PostURL(w, r)
	case http.MethodGet:
		handler.GetURL(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
