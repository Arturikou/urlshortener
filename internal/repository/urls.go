package repository

import (
	"errors"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/repository/file"
	"io"
	"sync"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("id already exists")

type Store struct {
	mu       sync.RWMutex
	urls     map[string]string
	producer *file.Producer
}

func NewStore(fileName string) (*Store, error) {
	store := &Store{
		urls: make(map[string]string),
	}
	if fileName == "" {
		return store, nil
	}

	consumer, err := file.NewConsumer(fileName)
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}
	defer consumer.Close()

	for {
		urlData, err := consumer.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read urlData: %w", err)
		}

		store.urls[urlData.ShortURL] = urlData.OriginalURL
	}

	producer, err := file.NewProducer(fileName)
	if err != nil {
		return nil, fmt.Errorf("create producer: %w", err)
	}

	store.producer = producer

	return store, nil
}

func (s *Store) Save(urlData *models.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.urls[urlData.ShortURL]; ok {
		return ErrAlreadyExists
	}

	if s.producer != nil {
		if err := s.producer.WriteEvent(urlData); err != nil {
			return fmt.Errorf("file save: %w", err)
		}
	}

	s.urls[urlData.ShortURL] = urlData.OriginalURL

	return nil
}

func (s *Store) Get(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return "", fmt.Errorf("get shortener with id %s: %w", id, ErrNotFound)
	}
	return url, nil
}
