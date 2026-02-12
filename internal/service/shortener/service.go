package shortener

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"go.uber.org/zap"
	"math/big"
)

const defaultMaxRetries = 10

//go:generate mockery
type Repository interface {
	Save(ctx context.Context, record *models.URLData) error
	Get(ctx context.Context, alias string) (string, error)
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

func (u *Service) AddURL(ctx context.Context, originalURL string) (string, error) {
	for i := 0; i < defaultMaxRetries; i++ {
		alias, err := generateAlias()
		if err != nil {
			return "", fmt.Errorf("failed to generate alias: %w", err)
		}

		urlData := &models.URLData{
			OriginalURL: originalURL,
			ShortURL:    alias,
		}

		err = u.repo.Save(ctx, urlData)
		if err == nil {
			return alias, nil
		}

		if errors.Is(err, models.ErrAlreadyExists) {
			continue
		}

		return "", fmt.Errorf("can't save alias: %w", err)
	}

	return "", fmt.Errorf("failed to generate unique id after %d attempts", defaultMaxRetries)
}

func (u *Service) GetURL(ctx context.Context, alias string) (string, error) {
	originalURL, err := u.repo.Get(ctx, alias)
	if err != nil {
		return "", fmt.Errorf("can't get alias: %w", err)
	}

	return originalURL, nil
}

func generateAlias() (string, error) {
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
