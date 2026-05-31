package shortener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/Arturikou/urlshortener/internal/service/shortener/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestShortener_GetStats(t *testing.T) {
	errCount := errors.New("count failed")

	tests := []struct {
		name      string
		setup     func(urlRepo *mocks.MockURLRepo, userRepo *mocks.MockUserURLRepo)
		wantStats models.Stats
		wantErr   bool
	}{
		{
			name: "success",
			setup: func(urlRepo *mocks.MockURLRepo, userRepo *mocks.MockUserURLRepo) {
				urlRepo.EXPECT().CountUrls(mock.Anything).Return(int64(10), nil)
				userRepo.EXPECT().CountUsers(mock.Anything).Return(int64(3), nil)
			},
			wantStats: models.Stats{URLsCount: 10, UserCount: 3},
			wantErr:   false,
		},
		{
			name: "urls count fails",
			setup: func(urlRepo *mocks.MockURLRepo, userRepo *mocks.MockUserURLRepo) {
				urlRepo.EXPECT().CountUrls(mock.Anything).Return(int64(0), errCount)
				userRepo.EXPECT().CountUsers(mock.Anything).Return(int64(3), nil).Maybe()
			},
			wantStats: models.Stats{},
			wantErr:   true,
		},
		{
			name: "users count fails",
			setup: func(urlRepo *mocks.MockURLRepo, userRepo *mocks.MockUserURLRepo) {
				urlRepo.EXPECT().CountUrls(mock.Anything).Return(int64(10), nil).Maybe()
				userRepo.EXPECT().CountUsers(mock.Anything).Return(int64(0), errCount)
			},
			wantStats: models.Stats{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlRepo := mocks.NewMockURLRepo(t)
			userRepo := mocks.NewMockUserURLRepo(t)
			tt.setup(urlRepo, userRepo)

			s := shortener.New(urlRepo, userRepo, nil, nil, zap.NewNop().Sugar())

			got, err := s.GetStats(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, models.Stats{}, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStats, got)
		})
	}
}
