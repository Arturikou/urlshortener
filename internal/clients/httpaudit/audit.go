// Package httpaudit provides an HTTP Audit client and methods to send events
package httpaudit

import (
	"encoding/json"
	"net/http"

	am "github.com/Arturikou/urlshortener/internal/managers/audit"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

type Client struct {
	client  *retryablehttp.Client
	address string
	logger  *zap.SugaredLogger
}

// New creates a new retryable HTTP Audit client.
func New(address string, logger *zap.SugaredLogger) *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = nil

	return &Client{
		address: address,
		logger:  logger,
		client:  retryClient,
	}
}

// Notify sends an event to a configured address via an HTTP POST request.
func (c *Client) Notify(event am.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		c.logger.Errorw("failed to marshal event", "error", err)
		return
	}

	request, err := retryablehttp.NewRequest(http.MethodPost, c.address, data)
	if err != nil {
		c.logger.Errorw("failed to create http request", "error", err)
		return
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		c.logger.Errorw("failed to send http request", "error", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		c.logger.Errorw("httpaudit event failed", "status", response.Status)
	}
}
