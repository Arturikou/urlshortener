package filestore

import (
	"context"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
)

func (fs *FileStore) Save(ctx context.Context, urlData *models.URLData) error {
	if err := fs.producer.WriteEvent(urlData); err != nil {
		return fmt.Errorf("file save: %w", err)
	}

	return fs.memoryStore.Save(ctx, urlData)
}

func (fs *FileStore) Get(ctx context.Context, id string) (string, error) {
	return fs.memoryStore.Get(ctx, id)
}
