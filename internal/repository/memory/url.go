package memory

import (
	"context"
	"github.com/Arturikou/urlshortener/internal/models"
)

func (s *Store) Save(ctx context.Context, urlData *models.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.urls[urlData.ShortURL]; ok {
		return models.ErrAlreadyExists
	}

	s.urls[urlData.ShortURL] = urlData.OriginalURL
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.urls[id]
	if !ok {
		return "", models.ErrNotFound
	}
	return url, nil
}
