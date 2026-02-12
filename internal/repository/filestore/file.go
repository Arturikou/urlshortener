package filestore

import (
	"fmt"
	"github.com/Arturikou/urlshortener/internal/repository/memory"
	"go.uber.org/zap"
	"io"
)

type FileStore struct {
	memoryStore *memory.Store
	producer    *Producer
	fileName    string
	logger      *zap.SugaredLogger
}

func NewFileStore(
	fileName string,
	memoryStore *memory.Store,
	logger *zap.SugaredLogger,
) (*FileStore, error) {
	producer, err := NewProducer(fileName)
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

func (fs *FileStore) Close() error {
	return fs.producer.Close()
}

func (fs *FileStore) Load() error {
	consumer, err := NewConsumer(fs.fileName)
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

		fs.memoryStore.Load(urlData.ShortURL, urlData.OriginalURL)
	}
	return nil
}
