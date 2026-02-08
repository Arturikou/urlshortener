package repository

import (
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/repository/file"
	"go.uber.org/zap"
	"io"
)

type FileStore struct {
	memoryStore *MemoryStore
	producer    *file.Producer
	fileName    string
	logger      *zap.SugaredLogger
}

func NewFileStore(
	fileName string,
	memoryStore *MemoryStore,
	logger *zap.SugaredLogger,
) (*FileStore, error) {
	producer, err := file.NewProducer(fileName)
	if err != nil {
		return nil, fmt.Errorf("create producer: %w", err)
	}

	return &FileStore{
		memoryStore: memoryStore,
		producer:    producer,
		fileName:    fileName,
		logger:      logger,
	}, nil
}

func (fs *FileStore) Save(urlData *models.URLData) error {
	if err := fs.producer.WriteEvent(urlData); err != nil {
		return fmt.Errorf("file save: %w", err)
	}

	return fs.memoryStore.Save(urlData)
}

func (fs *FileStore) Get(id string) (string, error) {
	return fs.memoryStore.Get(id)
}

func (fs *FileStore) Close() error {
	return fs.producer.Close()
}

func (fs *FileStore) Load() error {
	consumer, err := file.NewConsumer(fs.fileName)
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}
	defer consumer.Close()

	for {
		urlData, err := consumer.ReadEvent()
		if err == io.EOF {
			break
		}

		if err != nil {
			fs.logger.Errorf("Failed to decode line in file: %v", err)
			continue
		}

		fs.memoryStore.urls[urlData.ShortURL] = urlData.OriginalURL
	}
	return nil
}
