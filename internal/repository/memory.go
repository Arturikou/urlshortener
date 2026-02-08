package repository

import (
	"errors"
	"github.com/Arturikou/urlshortener/internal/models"
	"sync"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("id already exists")

type MemoryStore struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls: make(map[string]string),
	}
}

func (s *MemoryStore) Save(urlData *models.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.urls[urlData.ShortURL]; ok {
		return ErrAlreadyExists
	}

	s.urls[urlData.ShortURL] = urlData.OriginalURL
	return nil
}

func (s *MemoryStore) Get(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}
