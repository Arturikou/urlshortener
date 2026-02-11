package handlers

import (
	"context"
	"net/http"
	"time"
)

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		h.logger.Error("database repository is not initialized")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.repo.Ping(ctx); err != nil {
		h.logger.Errorf("database ping failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
