package router

import (
	"bytes"
	"encoding/json"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testResponse struct {
	StatusCode int
	Headers    http.Header
	Body       string
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) testResponse {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	if method == http.MethodPost {
		if strings.Contains(path, "shorten") {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "text/plain")
		}
	}

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return testResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(respBody),
	}
}

func TestRouter_AddURL(t *testing.T) {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	mockService := new(mocks.MockURLService)
	mockService.On("AddURL", "https://practicum.yandex.ru").Return("EwHXdJfB", nil)
	mockRepo := mocks.NewMockRepo(t)

	logger := zap.NewNop()
	h := handlers.New(mockService, mockRepo, cfg, logger.Sugar())
	ts := httptest.NewServer(New(h, logger))
	defer ts.Close()

	resp := testRequest(t, ts, http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru"))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "http://localhost:8080/EwHXdJfB", resp.Body)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "text/plain")
	mockService.AssertExpectations(t)
}

func TestRouter_GetURL(t *testing.T) {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	mockService := new(mocks.MockURLService)
	mockService.On("GetURL", "EwHXdJfB").
		Return("https://practicum.yandex.ru", nil)
	mockRepo := mocks.NewMockRepo(t)

	logger := zap.NewNop()
	h := handlers.New(mockService, mockRepo, cfg, logger.Sugar())
	ts := httptest.NewServer(New(h, logger))
	defer ts.Close()

	resp := testRequest(t, ts, http.MethodGet, "/EwHXdJfB", nil)

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, "https://practicum.yandex.ru", resp.Headers.Get("Location"))
	mockService.AssertExpectations(t)
}

func TestRouter_ShortenURL(t *testing.T) {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	mockService := new(mocks.MockURLService)
	mockService.On("AddURL", mock.Anything).
		Return("EwHXdJfB", nil).Once()
	mockRepo := mocks.NewMockRepo(t)

	logger := zap.NewNop()
	h := handlers.New(mockService, mockRepo, cfg, logger.Sugar())
	ts := httptest.NewServer(New(h, logger))
	defer ts.Close()

	bodyData, _ := json.Marshal(handlers.ShortenReq{URL: "https://practicum.yandex.ru"})
	resp := testRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewReader(bodyData))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var actualResp handlers.ShortenResp
	err := json.Unmarshal([]byte(resp.Body), &actualResp)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/EwHXdJfB", actualResp.Result)

	mockService.AssertExpectations(t)
}
