package filestore

import (
	"context"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
)

func (fs *FileStore) InsertOrGetAlias(ctx context.Context, urlData models.URLData) (models.UpsertResult, error) {
	upsertResult, err := fs.memoryStore.InsertOrGetAlias(ctx, urlData)
	if err != nil {
		return models.UpsertResult{}, err
	}

	if !upsertResult.IsInsert {
		return upsertResult, nil
	}

	if err = fs.producer.WriteEvent(&urlData); err != nil {
		_ = fs.memoryStore.Delete(ctx, &urlData)

		return models.UpsertResult{}, fmt.Errorf("filestore fix record: %w", err)
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
