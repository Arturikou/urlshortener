package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Arturikou/urlshortener/internal/models"
	"net/http"
	"net/url"
)

type ShortenReq struct {
	URL string `json:"url"`
}

type ShortenResp struct {
	Result string `json:"result"`
}

func (h *Handlers) Shorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenReq

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Debugw("error decoding request body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	alias, err := h.urlService.AddURL(r.Context(), req.URL)
	isConflict := errors.Is(err, models.ErrURLAlreadyExists)

	if err != nil && !isConflict {
		h.logger.Errorw("error add shortening URL", "url", req.URL, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortURL, err := url.JoinPath(h.cfg.BaseAddr, alias)
	if err != nil {
		h.logger.Errorw("failed to build short URL path", "base", h.cfg.BaseAddr, "alias", alias, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := ShortenResp{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if isConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Errorw("error encoding response", "url", req.URL, "error", err)
		return
	}
}
