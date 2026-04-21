package memory

import (
	"context"
	"fmt"

	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/google/uuid"
)

func (s *Store) AddURL(_ context.Context, data models.URLData) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.urlToAlias[data.OriginalURL]; ok {
		return 0, fmt.Errorf("url %q: %w", data.OriginalURL, models.ErrURLAlreadyShorted)
	}

	if _, ok := s.aliasToURL[data.Alias]; ok {
		return 0, fmt.Errorf("alias %q: %w", data.Alias, models.ErrAliasAlreadyExists)
	}

	s.counter++

	s.urlToAlias[data.OriginalURL] = models.URLRecord{
		ID:    s.counter,
		Alias: data.Alias,
		URL:   data.OriginalURL,
	}
	s.aliasToURL[data.Alias] = data.OriginalURL

	return s.counter, nil
}

func (s *Store) SaveBatch(_ context.Context, data []*models.ShortenBatch) ([]*models.ShortenBatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for idx, v := range data {
		if existingAlias, ok := s.urlToAlias[v.OriginalURL]; ok {
			v.Alias = existingAlias.Alias
			continue
		}

		if _, ok := s.aliasToURL[v.Alias]; ok {
			return data[idx:], nil
		}

		s.urlToAlias[v.OriginalURL] = models.URLRecord{
			ID:    s.counter,
			Alias: v.Alias,
		}

		s.aliasToURL[v.Alias] = v.OriginalURL
	}
	return nil, nil
}

func (s *Store) GetByAlias(_ context.Context, alias string) (models.URLRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.aliasToURL[alias]
	if !ok {
		return models.URLRecord{}, models.ErrNotFound
	}

	record, ok := s.urlToAlias[url]
	if !ok {
		return models.URLRecord{}, models.ErrNotFound
	}

	return record, nil
}

func (s *Store) GetByURL(_ context.Context, url string) (models.URLRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.urlToAlias[url]
	if !ok {
		return models.URLRecord{}, models.ErrNotFound
	}

	return record, nil
}

func (s *Store) Delete(_ context.Context, urlData *models.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.urlToAlias, urlData.OriginalURL)
	delete(s.aliasToURL, urlData.Alias)

	return nil
}

func (s *Store) AddUserURL(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}

func (s *Store) GetUserURLs(_ context.Context, _ uuid.UUID) ([]models.UserUrls, error) {
	return nil, nil
}

func (s *Store) DeleteURLs(_ context.Context, _ uuid.UUID, _ []string) error {
	return nil
}
