package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/Arturikou/urlshortener/internal/middleware"
)

type UserUrlsResp struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h *Handlers) UserUrls(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("failed to get userID from context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userURLs, err := h.urlService.GetUserURLs(r.Context(), userID)
	if err != nil {
		h.logger.Errorw("failed to get user URLs", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(userURLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]UserUrlsResp, len(userURLs))
	for idx, userURL := range userURLs {
		shortURL, err := url.JoinPath(h.cfg.BaseAddr, userURL.Alias)
		if err != nil {
			h.logger.Errorw("failed to build short URL path", "base", h.cfg.BaseAddr, "alias", userURL.Alias, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		resp[idx] = UserUrlsResp{
			ShortURL:    shortURL,
			OriginalURL: userURL.OriginalURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Errorw("error encoding response", "error", err)
		return
	}
}
