package memory

import (
	"context"
	"github.com/Arturikou/urlshortener/internal/models"
)

func (s *Store) InsertOrGetAlias(ctx context.Context, urlData models.URLData) (models.UpsertResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if actualAlias, ok := s.urlToAlias[urlData.OriginalURL]; ok {
		return models.UpsertResult{
			Alias:    actualAlias,
			IsInsert: false,
		}, nil
	}

	if _, ok := s.aliasToURL[urlData.Alias]; ok {
		return models.UpsertResult{}, models.ErrAliasAlreadyExists
	}

	s.urlToAlias[urlData.OriginalURL] = urlData.Alias
	s.aliasToURL[urlData.Alias] = urlData.OriginalURL

	return models.UpsertResult{
		Alias:    urlData.Alias,
		IsInsert: true,
	}, nil
}

func (s *Store) SaveBatch(ctx context.Context, data []*models.ShortenBatch) ([]*models.ShortenBatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for idx, v := range data {
		if existingAlias, ok := s.urlToAlias[v.OriginalURL]; ok {
			v.Alias = existingAlias
			continue
		}

		if _, ok := s.aliasToURL[v.Alias]; ok {
			return data[idx:], nil
		}

		s.urlToAlias[v.OriginalURL] = v.Alias
		s.aliasToURL[v.Alias] = v.OriginalURL
	}
	return nil, nil
}

func (s *Store) GetURLByAlias(ctx context.Context, alias string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.aliasToURL[alias]
	if !ok {
		return "", models.ErrNotFound
	}

	return url, nil
}

func (s *Store) Delete(ctx context.Context, urlData *models.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.urlToAlias, urlData.OriginalURL)
	delete(s.aliasToURL, urlData.Alias)

	return nil
}
