package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/handlers/config"
	"github.com/Arturikou/urlshortener/internal/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_Shorten(t *testing.T) {
	cfg := config.Config{BaseAddr: "http://localhost:8080"}
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
			name: "success",
			body: handlers.ShortenReq{
				URL: "https://practicum.yandex.ru",
			},
			setupService: func(m *mocks.MockURLService) {
				m.On("AddURL", mock.Anything).
					Return("/EwHXdJfB", nil).Once()
			},
			want: want{
				code:        http.StatusCreated,
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
				m.On("AddURL", mock.Anything).
					Return("", errors.New("error")).Once()
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
			test.setupService(mockService)

			h := handlers.New(mockService, cfg, logger)
			recorder := httptest.NewRecorder()

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(test.body)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", &buf)
			h.Shorten(recorder, request)

			assert.Equal(t, test.want.code, recorder.Code)
			assert.Equal(t, test.want.contentType, recorder.Header().Get("Content-Type"))

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
