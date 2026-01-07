package router

import (
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/handlers/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type URLServiceMock struct {
	mock.Mock
}

func (m *URLServiceMock) AddURL(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *URLServiceMock) GetURL(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

type testResponse struct {
	StatusCode int
	Headers    http.Header
	Body       string
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) testResponse {
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
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
	cfg := config.Config{
		BaseAddr: "http://localhost:8080",
	}

	mockService := new(URLServiceMock)
	mockService.
		On("AddURL", "https://practicum.yandex.ru").
		Return("EwHXdJfB", nil)

	h := handlers.New(mockService, cfg)
	ts := httptest.NewServer(New(h))
	defer ts.Close()

	resp := testRequest(
		t,
		ts,
		http.MethodPost,
		"/",
		strings.NewReader("https://practicum.yandex.ru"),
	)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "http://localhost:8080/EwHXdJfB", resp.Body)
	assert.Equal(t, "text/plain", resp.Headers.Get("Content-Type"))

	mockService.AssertExpectations(t)
}

func TestRouter_GetURL(t *testing.T) {
	cfg := config.Config{
		BaseAddr: "http://localhost:8080",
	}

	mockService := new(URLServiceMock)
	mockService.
		On("GetURL", "EwHXdJfB").
		Return("https://practicum.yandex.ru", nil)

	h := handlers.New(mockService, cfg)
	ts := httptest.NewServer(New(h))
	defer ts.Close()

	resp := testRequest(
		t,
		ts,
		http.MethodGet,
		"/EwHXdJfB",
		nil,
	)

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, "https://practicum.yandex.ru", resp.Headers.Get("Location"))
	mockService.AssertExpectations(t)
}
