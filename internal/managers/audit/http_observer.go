// Package audit provides an Audit observer that sends the audit events to an HTTP client
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

// NewHTTPObserver creates a new HTTP Audit observer
func NewHTTPObserver(
	client HTTPClient,
	logger *zap.SugaredLogger,
) *HTTPObserver {
	return &HTTPObserver{
		client: client,
		logger: logger,
	}
}

// Notify sends the event to the Audit HTTP client
func (h *HTTPObserver) Notify(event Event) {
	h.client.Notify(event)
}
