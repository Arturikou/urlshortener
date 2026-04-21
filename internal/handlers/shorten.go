package handlers

import (
	"encoding/json"
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

	addResult, err := h.urlService.AddURL(r.Context(), req.URL)

	if err != nil {
		h.logger.Errorw("error add shortening URL", "url", req.URL, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortURL, err := url.JoinPath(h.cfg.BaseAddr, addResult.Alias)
	if err != nil {
		h.logger.Errorw("failed to build short URL path", "base", h.cfg.BaseAddr, "alias", addResult.Alias, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := ShortenResp{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if addResult.IsInsert {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Errorw("error encoding response", "url", req.URL, "error", err)
		return
	}
}
