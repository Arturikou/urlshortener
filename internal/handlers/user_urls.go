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

// UserUrls godoc
//
//	@Summary		Get user's URLs
//	@Description	Returns all URLs shortened by the authenticated user
//	@Tags			user
//	@Produce		json
//	@Security		CookieAuth
//	@Success		200	{array}		UserUrlsResp
//	@Success		204
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/api/user/urls [get]
func (h *Handlers) UserUrls(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
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

// DeleteUserURLs godoc
//
//	@Summary		Delete user's URLs
//	@Description	Asynchronously deletes the given short URL aliases belonging to the authenticated user
//	@Tags			user
//	@Accept			json
//	@Security		CookieAuth
//	@Param			aliases	body	[]string	true	"List of aliases to delete"
//	@Success		202
//	@Failure		400	{string}	string	"Bad request or empty list"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/api/user/urls [delete]
func (h *Handlers) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		h.logger.Error("failed to get userID from context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var aliases []string
	err = json.NewDecoder(r.Body).Decode(&aliases)
	if err != nil {
		h.logger.Errorw("error decoding request body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(aliases) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	h.deleteEnqueuer.EnqueueDelete(userID, aliases)

	w.WriteHeader(http.StatusAccepted)
}
