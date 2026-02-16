package memory

import (
	"sync"
)

type Store struct {
	mu         sync.RWMutex
	aliasToURL map[string]string
	urlToAlias map[string]string
}

func NewMemoryStore() *Store {
	return &Store{
		aliasToURL: make(map[string]string),
		urlToAlias: make(map[string]string),
	}
}

func (s *Store) Load(alias, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aliasToURL[alias] = originalURL
	s.urlToAlias[originalURL] = alias
}
