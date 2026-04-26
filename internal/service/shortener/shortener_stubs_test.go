package shortener

import (
	"context"

	"github.com/Arturikou/urlshortener/internal/managers/audit"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/google/uuid"
)

// --- stubs ---

type stubURLRepo struct{}

func (s *stubURLRepo) AddURL(_ context.Context, _ models.URLData) (int64, error) {
	return 1, nil
}

func (s *stubURLRepo) GetByAlias(_ context.Context, _ string) (models.URLRecord, error) {
	return models.URLRecord{ID: 1, Alias: "test", URL: "https://practicum.yandex.com"}, nil
}

func (s *stubURLRepo) GetByURL(_ context.Context, _ string) (models.URLRecord, error) {
	return models.URLRecord{}, models.ErrNotFound
}

func (s *stubURLRepo) SaveBatch(_ context.Context, b []models.ShortenBatch) ([]models.ShortenBatch, error) {
	return nil, nil
}

func (s *stubURLRepo) DeleteURLs(_ context.Context, _ uuid.UUID, _ []string) error {
	return nil
}

type stubUserURLRepo struct{}

func (s *stubUserURLRepo) AddUserURL(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}

func (s *stubUserURLRepo) GetUserURLs(_ context.Context, _ uuid.UUID) ([]models.UserUrls, error) {
	return nil, nil
}

type stubTransactor struct{}

func (s *stubTransactor) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type stubAuditManager struct{}

func (n *stubAuditManager) NotifyAll(_ context.Context, _ audit.Event) {}
