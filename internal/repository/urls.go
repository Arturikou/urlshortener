package repository

import (
	"errors"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	urls map[string]string
}

func NewStore() *Store {
	return &Store{
		urls: make(map[string]string),
	}
}

func (s *Store) Save(id, url string) error {
	s.urls[id] = url
	return nil
}

func (s *Store) Get(id string) (string, error) {
	url, ok := s.urls[id]

	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}
