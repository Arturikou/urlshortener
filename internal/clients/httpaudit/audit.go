package httpaudit

import (
	"bytes"
	"encoding/json"
	"net/http"

	am "github.com/Arturikou/urlshortener/internal/managers/audit"
	"go.uber.org/zap"
)

type Client struct {
	httpClient http.Client
	address    string
	logger     *zap.SugaredLogger
}

func New(
	address string,
	logger *zap.SugaredLogger,
) *Client {
	return &Client{
		address:    address,
		logger:     logger,
		httpClient: http.Client{},
	}
}

func (c *Client) Notify(event am.Event) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(event); err != nil {
		c.logger.Errorw("failed to encode event", "error", err)
		return
	}
	request, err := http.NewRequest(http.MethodPost, c.address, &buf)
	if err != nil {
		c.logger.Errorw("failed to create http request", "error", err)
		return
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		c.logger.Errorw("failed to send http request", "error", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		c.logger.Errorw("httpaudit event failed", "status", response.Status)
	}
}
