package storage

import (
	"context"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/repository/filestore"
	"github.com/Arturikou/urlshortener/internal/repository/memory"

	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/repository/postgres"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"go.uber.org/zap"
)

type Storage struct {
	URLRepo shortener.Repository
	Pinger  Pinger
	Closer  func()
}

func New(
	ctx context.Context,
	cfg *config.Config,
	log *zap.SugaredLogger,
) (*Storage, error) {
	if cfg.Database.DSN != "" {
		db, err := postgres.New(ctx, cfg.Database.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		return &Storage{
			URLRepo: db,
			Pinger:  db,
			Closer:  db.Close,
		}, nil
	}

	if cfg.FileStoragePath != "" {
		memoryStore := memory.NewMemoryStore()
		fileStore, err := filestore.NewFileStore(cfg.FileStoragePath, memoryStore, log)
		if err != nil {
			return nil, err
		}
		if err = fileStore.Load(); err != nil {
			return nil, fmt.Errorf("file storage load: %w", err)
		}
		return &Storage{
			URLRepo: fileStore,
			Closer:  func() { _ = fileStore.Close() },
		}, nil
	}

	memoryStore := memory.NewMemoryStore()
	return &Storage{
		URLRepo: memoryStore,
	}, nil
}
