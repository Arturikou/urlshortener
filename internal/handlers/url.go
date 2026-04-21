package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/go-chi/chi/v5"
)

// AddURL godoc
//
//	@Summary		Shorten URL (plain text)
//	@Description	Accepts a plain-text URL in the body and returns the shortened URL
//	@Tags			urls
//	@Accept			plain
//	@Produce		plain
//	@Param			url	body		string	true	"Original URL"
//	@Success		201	{string}	string	"Short URL"
//	@Failure		400	{string}	string	"Empty body"
//	@Failure		409	{string}	string	"URL already shortened"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/ [post]
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
	addResult, err := h.urlService.AddURL(r.Context(), originalURL)
	if err != nil {
		h.logger.Errorw("failed to add URL", "url", originalURL, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortURL, err := url.JoinPath(h.cfg.BaseAddr, addResult.Alias)
	if err != nil {
		h.logger.Errorw("failed to build short URL path", "base", h.cfg.BaseAddr, "alias", addResult.Alias, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	if addResult.IsInsert {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		h.logger.Errorw("failed to write body", "url", originalURL, "error", err)
	}
}

// GetURL godoc
//
//	@Summary		Redirect to original URL
//	@Description	Redirects to the original URL by short alias
//	@Tags			urls
//	@Param			alias	path	string	true	"Short URL alias"
//	@Success		307
//	@Failure		400	{string}	string	"Alias is required"
//	@Failure		404	{string}	string	"Not found"
//	@Failure		410	{string}	string	"URL has been deleted"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{alias} [get]
func (h *Handlers) GetURL(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")

	if alias == "" {
		http.Error(w, "alias is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.urlService.GetURL(r.Context(), alias)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "shortener not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, models.ErrURLDeleted) {
			http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
		}

		h.logger.Errorw("database error during GetURL", "alias", alias, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
