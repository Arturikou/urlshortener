package repository

import (
	"errors"
	"fmt"
	"sync"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("id already exists")

type Store struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewStore() *Store {
	return &Store{
		urls: make(map[string]string),
	}
}

func (s *Store) Save(id, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.urls[id]; ok {
		return ErrAlreadyExists
	}

	s.urls[id] = url
	return nil
}

func (s *Store) Get(id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return "", fmt.Errorf("get url with id %s: %w", id, ErrNotFound)
	}
	return url, nil
}
