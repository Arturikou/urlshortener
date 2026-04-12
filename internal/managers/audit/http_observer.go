package audit

import (
	"go.uber.org/zap"
)

type HTTPClient interface {
	Notify(event Event)
}

type HTTPObserver struct {
	client HTTPClient
	logger *zap.SugaredLogger
}

func NewHTTPObserver(
	client HTTPClient,
	logger *zap.SugaredLogger,
) *HTTPObserver {
	return &HTTPObserver{
		client: client,
		logger: logger,
	}
}

func (h *HTTPObserver) Notify(event Event) {
	h.client.Notify(event)
}
