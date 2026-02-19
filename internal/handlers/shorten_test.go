package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/handlers/mocks"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	pinger "github.com/Arturikou/urlshortener/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_Shorten(t *testing.T) {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	logger := zap.NewNop().Sugar()

	type want struct {
		code        int
		contentType string
		result      string
	}

	tests := []struct {
		name         string
		body         any
		setupService func(m *mocks.MockURLService)
		want         want
	}{
		{
			name: "success created",
			body: handlers.ShortenReq{
				URL: "https://practicum.yandex.ru",
			},
			setupService: func(m *mocks.MockURLService) {
				m.On("AddURL", mock.Anything, "https://practicum.yandex.ru").
					Return(shortener.AddURLResult{Alias: "EwHXdJfB", IsInsert: true}, nil).Once()
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
				result:      "http://localhost:8080/EwHXdJfB",
			},
		},
		{
			name: "success conflict",
			body: handlers.ShortenReq{
				URL: "https://practicum.yandex.ru",
			},
			setupService: func(m *mocks.MockURLService) {
				m.On("AddURL", mock.Anything, "https://practicum.yandex.ru").
					Return(shortener.AddURLResult{Alias: "EwHXdJfB", IsInsert: false}, nil).Once()
			},
			want: want{
				code:        http.StatusConflict,
				contentType: "application/json",
				result:      "http://localhost:8080/EwHXdJfB",
			},
		},
		{
			name:         "Invalid JSON",
			body:         `{"url": "broken"`,
			setupService: func(m *mocks.MockURLService) {},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "empty url",
			body: handlers.ShortenReq{
				URL: "",
			},
			setupService: func(m *mocks.MockURLService) {},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "error from db",
			body: handlers.ShortenReq{
				URL: "https://practicum.yandex.ru",
			},
			setupService: func(m *mocks.MockURLService) {
				m.On("AddURL", mock.Anything, mock.Anything).
					Return(shortener.AddURLResult{}, errors.New("internal error")).Once()
			},
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(mocks.MockURLService)
			mockPinger := new(pinger.MockPinger)
			test.setupService(mockService)

			h := handlers.New(mockService, mockPinger, cfg, logger)
			recorder := httptest.NewRecorder()

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(test.body)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", &buf)
			h.Shorten(recorder, request)

			assert.Equal(t, test.want.code, recorder.Code)
			assert.Contains(t, recorder.Header().Get("Content-Type"), test.want.contentType)

			if test.want.result != "" {
				var res handlers.ShortenResp
				err := json.NewDecoder(recorder.Body).Decode(&res)
				require.NoError(t, err)
				assert.Equal(t, test.want.result, res.Result)
			}
			mockService.AssertExpectations(t)
		})
	}
}
