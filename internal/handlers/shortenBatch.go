package handlers

import (
	"encoding/json"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"net/http"
	"net/url"
)

type ShortenBatchReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResp struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (h *Handlers) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req []ShortenBatchReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Errorw("error decoding request body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	shortenBatch := make([]*shortener.ShortenBatch, len(req))
	for idx, reqItem := range req {
		shortenBatch[idx] = &shortener.ShortenBatch{
			CorrelationID: reqItem.CorrelationID,
			OriginalURL:   reqItem.OriginalURL,
		}
	}

	results, err := h.urlService.AddURLs(r.Context(), shortenBatch)
	if err != nil {
		h.logger.Errorw("failed to add batch URLs", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortenBatchResp := make([]ShortenBatchResp, len(results))
	for idx, result := range results {
		shortURL, err := url.JoinPath(h.cfg.BaseAddr, result.Alias)
		if err != nil {
			h.logger.Errorw("failed to build short URL path", "base", h.cfg.BaseAddr, "alias", result.Alias, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		shortenBatchResp[idx] = ShortenBatchResp{
			CorrelationID: result.CorrelationID,
			ShortURL:      shortURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(shortenBatchResp)
	if err != nil {
		h.logger.Errorw("error encoding response", "error", err)
		return
	}

}
