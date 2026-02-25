package filestore

import (
	"context"
	"fmt"

	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/google/uuid"
)

func (fs *FileStore) AddURL(ctx context.Context, data models.URLData) (int64, error) {
	upsertResult, err := fs.memoryStore.AddURL(ctx, data)
	if err != nil {
		return 0, fmt.Errorf("filestore add url: %w", err)
	}

	if err = fs.producer.WriteEvent(&data); err != nil {
		_ = fs.memoryStore.Delete(ctx, &data)

		return 0, fmt.Errorf("filestore write event: %w", err)
	}

	return upsertResult, nil
}

func (fs *FileStore) SaveBatch(ctx context.Context, data []*models.ShortenBatch) ([]*models.ShortenBatch, error) {
	remaining, err := fs.memoryStore.SaveBatch(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("fail to memory save: %w", err)
	}

	savedCount := len(data) - len(remaining)
	if savedCount > 0 {
		if err = fs.producer.WriteBatch(data[:savedCount]); err != nil {
			return nil, fmt.Errorf("fail to file save: %w", err)
		}
	}

	return remaining, nil
}

func (fs *FileStore) GetURLByAlias(ctx context.Context, id string) (string, error) {
	return fs.memoryStore.GetURLByAlias(ctx, id)
}

func (fs *FileStore) GetByURL(ctx context.Context, url string) (models.URLRecord, error) {
	return fs.memoryStore.GetByURL(ctx, url)
}

func (fs *FileStore) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error) {
	return fs.memoryStore.GetUserURLs(ctx, userID)
}

func (fs *FileStore) AddUserURL(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}
