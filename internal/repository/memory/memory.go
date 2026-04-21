package memory

import (
	"context"
	"sync"

	"github.com/Arturikou/urlshortener/internal/models"
)

type Store struct {
	mu         sync.RWMutex
	aliasToURL map[string]string
	urlToAlias map[string]models.URLRecord
	counter    int64
}

func NewMemoryStore() *Store {
	return &Store{
		aliasToURL: make(map[string]string),
		urlToAlias: make(map[string]models.URLRecord),
	}
}

func (s *Store) Load(alias, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aliasToURL[alias] = originalURL
	s.urlToAlias[originalURL] = models.URLRecord{
		ID:    s.counter,
		Alias: alias,
	}
}

func (s *Store) Transaction(_ context.Context, fn func(ctx context.Context) error) error {
	return fn(context.Background())
}
