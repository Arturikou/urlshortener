package memory

import (
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemoryStore() *Store {
	return &Store{
		urls: make(map[string]string),
	}
}

func (s *Store) Load(shortURL, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[shortURL] = originalURL
}
