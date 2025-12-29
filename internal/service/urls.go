package service

import "fmt"

type urlRepo interface {
	Save(id, url string) error
	Get(id string) (string, error)
}

type URLs struct {
	repo urlRepo
}

func New(repo urlRepo) *URLs {
	return &URLs{repo}
}

func (u *URLs) AddURL(url string) (string, error) {
	// временно пока не будет добавлена логика
	id := "EwHXdJfB"
	err := u.repo.Save(id, url)
	if err != nil {
		return "", fmt.Errorf(`can't save id': %w`, err)
	}

	return id, nil
}

func (u *URLs) GetURL(id string) (string, error) {
	originalURL, err := u.repo.Get(id)
	if err != nil {
		return "", fmt.Errorf(`can't get id': %w`, err)
	}
	return originalURL, nil
}
