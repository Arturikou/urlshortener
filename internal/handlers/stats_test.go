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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newStatsHandler(mockService *mocks.MockURLService) *handlers.Handlers {
	cfg := config.HandlersConfig{BaseAddr: "http://localhost:8080"}
	logger := zap.NewNop().Sugar()
	mockEnqueuer := new(mocks.MockDeleteEnqueuer)

	return handlers.New(mockService, mockEnqueuer, new(pinger.MockPinger), cfg, logger)
}

func TestHandlers_GetStats_Success(t *testing.T) {
	mockService := new(mocks.MockURLService)
	mockService.
		On("GetStats", mock.Anything).
		Return(models.Stats{URLsCount: 10, UserCount: 3}, nil).Once()

	h := newStatsHandler(mockService)

	request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	recorder := httptest.NewRecorder()

	h.GetStats(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Header().Get("Content-Type"), "application/json")

	var resp handlers.StatsResp
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	assert.Equal(t, handlers.StatsResp{Urls: 10, Users: 3}, resp)

	mockService.AssertExpectations(t)
}

func TestHandlers_GetStats_ServiceError(t *testing.T) {
	mockService := new(mocks.MockURLService)
	mockService.
		On("GetStats", mock.Anything).
		Return(models.Stats{}, errors.New("error")).Once()

	h := newStatsHandler(mockService)

	request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
	recorder := httptest.NewRecorder()

	h.GetStats(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	mockService.AssertExpectations(t)
}
