package url

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/repository"
	"math/big"
)

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
	for {
		id, err := generateID()
		if err != nil {
			return "", fmt.Errorf("failed to generate id: %w", err)
		}

		err = u.repo.Save(id, url)
		if err == nil {
			return id, nil
		}

		if errors.Is(err, repository.ErrAlreadyExists) {
			continue
		}

		return "", fmt.Errorf(`can't save id': %w`, err)
	}
}

func (u *Service) GetURL(id string) (string, error) {
	originalURL, err := u.repo.Get(id)
	if err != nil {
		return "", fmt.Errorf(`can't get id': %w`, err)
	}

	return originalURL, nil
}

func generateID() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	id := make([]byte, 10)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range id {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random int: %w", err)
		}
		id[i] = charset[n.Int64()]
	}

	return string(id), nil
}
