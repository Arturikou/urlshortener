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
	Save(ctx context.Context, record models.URLData) (string, error)
	Get(ctx context.Context, alias string) (string, error)
	SaveBatch(ctx context.Context, data []*ShortenBatch) ([]*ShortenBatch, error)
}

type Shortener struct {
	repo   Repository
	logger *zap.SugaredLogger
}

func New(repo Repository, logger *zap.SugaredLogger) *Shortener {
	return &Shortener{
		repo:   repo,
		logger: logger,
	}
}

type ShortenBatch struct {
	CorrelationID string
	OriginalURL   string
	Alias         string
}

func (s *Shortener) AddURL(ctx context.Context, originalURL string) (string, error) {
	for i := 0; i < defaultMaxRetries; i++ {
		alias, err := generateAlias()
		if err != nil {
			return "", fmt.Errorf("failed to generate alias: %w", err)
		}

		urlData := models.URLData{
			OriginalURL: originalURL,
			Alias:       alias,
		}

		actualAlias, err := s.repo.Save(ctx, urlData)
		if err == nil {
			return actualAlias, nil
		}

		if errors.Is(err, models.ErrURLAlreadyExists) {
			return actualAlias, err
		}

		if errors.Is(err, models.ErrAliasAlreadyExists) {
			continue
		}

		return "", fmt.Errorf("can't save alias: %w", err)
	}

	return "", fmt.Errorf("failed to generate unique id after %d attempts", defaultMaxRetries)
}

func (s *Shortener) AddURLs(ctx context.Context, batches []*ShortenBatch) ([]ShortenBatch, error) {
	batchSize := 1000

	for i := 0; i < len(batches); i += batchSize {
		end := i + batchSize
		if end > len(batches) {
			end = len(batches)
		}

		for _, item := range batches[i:end] {
			alias, err := generateAlias()
			if err != nil {
				return nil, fmt.Errorf("failed to generate alias: %w", err)
			}
			item.Alias = alias
		}

		currentBatch := batches[i:end]
		retries := 0

		for len(currentBatch) > 0 {
			remaining, err := s.repo.SaveBatch(ctx, currentBatch)
			if err != nil {
				return nil, fmt.Errorf("failed to save batch: %w", err)
			}

			if len(remaining) == 0 {
				break
			}

			if retries > defaultMaxRetries {
				return nil, fmt.Errorf("too many retries")
			}

			newAlias, err := generateAlias()
			if err != nil {
				return nil, fmt.Errorf("failed to regenerate alias: %w", err)
			}
			remaining[0].Alias = newAlias

			currentBatch = remaining
			retries++
		}
	}

	result := make([]ShortenBatch, len(batches))
	for idx, item := range batches {
		result[idx] = *item
	}

	return result, nil
}

func (s *Shortener) GetURL(ctx context.Context, alias string) (string, error) {
	originalURL, err := s.repo.Get(ctx, alias)
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
