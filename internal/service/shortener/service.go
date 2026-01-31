package shortener

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"math/big"
)

const defaultMaxRetries = 10

type Repository interface {
	Save(record *models.URLData) error
	Get(id string) (string, error)
}

type Service struct {
	repo   Repository
	logger *zap.SugaredLogger
}

func New(repo Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (u *Service) AddURL(originalURL string) (string, error) {
	for i := 0; i < defaultMaxRetries; i++ {
		id, err := generateID()
		if err != nil {
			return "", fmt.Errorf("failed to generate id: %w", err)
		}

		newUUID := uuid.New().String()
		urlData := &models.URLData{
			OriginalURL: originalURL,
			ShortURL:    id,
			UUID:        newUUID,
		}

		err = u.repo.Save(urlData)
		if err == nil {
			return id, nil
		}

		if errors.Is(err, repository.ErrAlreadyExists) {
			continue
		}

		return "", fmt.Errorf("can't save id: %w", err)
	}

	return "", fmt.Errorf("failed to generate unique id after %d attempts", defaultMaxRetries)
}

func (u *Service) GetURL(id string) (string, error) {
	originalURL, err := u.repo.Get(id)
	if err != nil {
		return "", fmt.Errorf("can't get id: %w", err)
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
