package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/handlers/mocks"
	"github.com/Arturikou/urlshortener/internal/models"
	pinger "github.com/Arturikou/urlshortener/internal/storage/mocks"
	"github.com/Arturikou/urlshortener/internal/userctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHandlers_UserUrls(t *testing.T) {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	logger := zap.NewNop().Sugar()

	userID := uuid.New()

	type mockData struct {
		returnURLs []models.UserUrls
		returnErr  error
	}

	type want struct {
		code        int
		contentType string
		body        []handlers.UserUrlsResp
	}

	tests := []struct {
		name        string
		userIDInCtx bool
		mock        mockData
		want        want
	}{
		{
			name:        "Success",
			userIDInCtx: true,
			mock: mockData{
				returnURLs: []models.UserUrls{
					{OriginalURL: "https://practicum.yandex.ru", Alias: "EwHXdJfB"},
					{OriginalURL: "https://google.com", Alias: "AbCdEfGh"},
				},
				returnErr: nil,
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				body: []handlers.UserUrlsResp{
					{ShortURL: "http://localhost:8080/EwHXdJfB", OriginalURL: "https://practicum.yandex.ru"},
					{ShortURL: "http://localhost:8080/AbCdEfGh", OriginalURL: "https://google.com"},
				},
			},
		},
		{
			name:        "No URLs",
			userIDInCtx: true,
			mock: mockData{
				returnURLs: []models.UserUrls{},
				returnErr:  nil,
			},
			want: want{
				code: http.StatusNoContent,
			},
		},
		{
			name:        "Service error",
			userIDInCtx: true,
			mock: mockData{
				returnURLs: nil,
				returnErr:  errors.New("db error"),
			},
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain",
			},
		},
		{
			name:        "No userID in context",
			userIDInCtx: false,
			want: want{
				code:        http.StatusInternalServerError,
				contentType: "text/plain",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := new(mocks.MockURLService)
			if test.userIDInCtx && test.mock.returnErr != nil || test.userIDInCtx && test.mock.returnURLs != nil {
				mockService.
					On("GetUserURLs", mock.Anything, userID).
					Return(test.mock.returnURLs, test.mock.returnErr).Once()
			}

			mockPinger := pinger.NewMockPinger(t)
			mockEnqueuer := new(mocks.MockDeleteEnqueuer)
			h := handlers.New(mockService, mockEnqueuer, mockPinger, cfg, logger)

			r := chi.NewRouter()
			r.Get("/api/user/urls", h.UserUrls)

			request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

			if test.userIDInCtx {
				ctx := userctx.WithUserID(request.Context(), userID)
				request = request.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			assert.Equal(t, test.want.code, w.Code)

			if test.want.contentType != "" {
				assert.Contains(t, w.Header().Get("Content-Type"), test.want.contentType)
			}

			if test.want.body != nil {
				var resp []handlers.UserUrlsResp
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, test.want.body, resp)
			}

			mockService.AssertExpectations(t)
		})
	}
}
