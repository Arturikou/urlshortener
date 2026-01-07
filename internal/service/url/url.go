package url

import "fmt"

type repo interface {
	Save(id, url string) error
	Get(id string) (string, error)
}

type Service struct {
	repo repo
}

func New(repo repo) *Service {
	return &Service{repo}
}

func (u *Service) AddURL(url string) (string, error) {
	// временно пока не будет добавлена логика
	id := "EwHXdJfB"
	err := u.repo.Save(id, url)
	if err != nil {
		return "", fmt.Errorf(`can't save id': %w`, err)
	}

	return id, nil
}

func (u *Service) GetURL(id string) (string, error) {
	originalURL, err := u.repo.Get(id)
	if err != nil {
		return "", fmt.Errorf(`can't get id': %w`, err)
	}
	return originalURL, nil
}
