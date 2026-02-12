package handlers_test

import (
	"errors"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/handlers/mocks"
	"github.com/Arturikou/urlshortener/internal/models"
	pinger "github.com/Arturikou/urlshortener/internal/storage/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlers_AddURL(t *testing.T) {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	logger := zap.NewNop().Sugar()

	type mockData struct {
		returnID  string
		returnErr error
	}

	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name string
		body string
		mock mockData
		want want
	}{
		{
			name: "Success",
			body: "https://practicum.yandex.ru",
			mock: mockData{
				returnID:  "EwHXdJfB",
				returnErr: nil,
			},
			want: want{
				code:        http.StatusCreated,
				response:    "http://localhost:8080/EwHXdJfB",
				contentType: "text/plain",
			},
		},
		{
			name: "Empty body",
			body: "",
			mock: mockData{
				returnID:  "EwHXdJfB",
				returnErr: nil,
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "empty body",
				contentType: "text/plain",
			},
		},
		{
			name: "Error from service",
			body: "https://practicum.yandex.ru",
			mock: mockData{
				returnID:  "",
				returnErr: errors.New("error"),
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "Internal Server Error",
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(mocks.MockURLService)
			mockService.
				On("AddURL", mock.Anything, test.body).
				Return(test.mock.returnID, test.mock.returnErr)

			mockPinger := pinger.NewMockPinger(t)

			h := handlers.New(mockService, mockPinger, cfg, logger)
			r := chi.NewRouter()
			r.Post("/", h.AddURL)

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			assert.Equal(t, test.want.code, w.Code)
			assert.Equal(t, test.want.response, strings.TrimSpace(w.Body.String()))
			assert.Contains(t, w.Header().Get("Content-Type"), test.want.contentType)
		})
	}
}

func TestHandlers_GetURL(t *testing.T) {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	logger := zap.NewNop().Sugar()

	type mockData struct {
		returnURL string
		returnErr error
	}

	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name  string
		alias string
		mock  mockData
		want  want
	}{
		{
			name:  "Success",
			alias: "EwHXdJfB",
			mock: mockData{
				returnURL: "https://practicum.yandex.ru/",
				returnErr: nil,
			},
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/",
			},
		},
		{
			name:  "Empty alias",
			alias: "",
			mock: mockData{
				returnURL: "https://practicum.yandex.ru/",
				returnErr: nil,
			},
			want: want{
				code:     http.StatusNotFound,
				location: "",
			},
		},
		{
			name:  "ID not found",
			alias: "DFSVDSFdd",
			mock: mockData{
				returnURL: "https://practicum.yandex.ru/",
				returnErr: models.ErrNotFound,
			},
			want: want{
				code:     http.StatusNotFound,
				location: "",
			},
		},
		{
			name:  "Error from service",
			alias: "EwHXdJfB",
			mock: mockData{
				returnURL: "",
				returnErr: errors.New("error"),
			},
			want: want{
				code:     http.StatusInternalServerError,
				location: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPinger := pinger.NewMockPinger(t)
			mockService := new(mocks.MockURLService)
			mockService.
				On("GetURL", mock.Anything, test.alias).
				Return(test.mock.returnURL, test.mock.returnErr)

			h := handlers.New(mockService, mockPinger, cfg, logger)
			r := chi.NewRouter()
			r.Get("/{alias}", h.GetURL)

			request := httptest.NewRequest(http.MethodGet, "/"+test.alias, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			assert.Equal(t, test.want.code, w.Code)
			assert.Equal(t, test.want.location, w.Header().Get("Location"))
		})
	}
}
