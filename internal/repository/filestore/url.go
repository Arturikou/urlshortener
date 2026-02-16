package filestore

import (
	"context"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
)

func (fs *FileStore) Save(ctx context.Context, urlData models.URLData) (string, error) {
	actualAlias, err := fs.memoryStore.Save(ctx, urlData)
	if err != nil {
		return "", err
	}

	if actualAlias != urlData.Alias {
		return actualAlias, nil
	}

	urlData.Alias = actualAlias

	if err := fs.producer.WriteEvent(&urlData); err != nil {
		_ = fs.memoryStore.Delete(ctx, &urlData)
		return "", fmt.Errorf("file save: %w", err)
	}

	return urlData.Alias, nil
}

func (fs *FileStore) SaveBatch(ctx context.Context, data []*shortener.ShortenBatch) ([]*shortener.ShortenBatch, error) {
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

func (fs *FileStore) Get(ctx context.Context, id string) (string, error) {
	return fs.memoryStore.Get(ctx, id)
}
