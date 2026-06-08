package handlers

import (
	"encoding/json"
	"net/http"
)

type StatsResp struct {
	Urls  int64 `json:"urls"`
	Users int64 `json:"users"`
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.urlService.GetStats(r.Context())
	if err != nil {
		h.logger.Errorw("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := StatsResp{
		Urls:  stats.URLsCount,
		Users: stats.UserCount,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		h.logger.Errorw("error encoding response", "error", err)
		return
	}
}
