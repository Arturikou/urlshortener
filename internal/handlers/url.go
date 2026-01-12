package handlers

import (
	"errors"
	"github.com/Arturikou/urlshortener/internal/repository"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (h *Handlers) AddURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	id, err := h.urlService.AddURL(originalURL)
	if err != nil {
		log.Printf("AddURL error: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortURL, _ := url.JoinPath(h.cfg.BaseAddr, id)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handlers) GetURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.urlService.GetURL(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "url not found", http.StatusNotFound)
		}

		log.Printf("GetURL error: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
