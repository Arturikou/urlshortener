package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"go.uber.org/zap"
)

const defaultMaxRetries = 10

//go:generate mockery
type Repository interface {
	InsertOrGetAlias(ctx context.Context, data models.URLData) (models.UpsertResult, error)
	GetURLByAlias(ctx context.Context, alias string) (string, error)
	SaveBatch(ctx context.Context, data []*models.ShortenBatch) ([]*models.ShortenBatch, error)
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

type AddURLResult struct {
	Alias    string
	IsInsert bool
}

func (s *Shortener) AddURL(ctx context.Context, originalURL string) (AddURLResult, error) {
	for i := 0; i < defaultMaxRetries; i++ {
		alias, err := generateAlias()
		if err != nil {
			return AddURLResult{}, fmt.Errorf("failed to generate alias: %w", err)
		}

		urlData := models.URLData{
			OriginalURL: originalURL,
			Alias:       alias,
		}

		upsertResult, err := s.repo.InsertOrGetAlias(ctx, urlData)
		if err == nil {
			return AddURLResult{
				Alias:    upsertResult.Alias,
				IsInsert: upsertResult.IsInsert,
			}, nil
		}

		if errors.Is(err, models.ErrAliasAlreadyExists) {
			continue
		}

		return AddURLResult{}, fmt.Errorf("can't save alias: %w", err)
	}

	return AddURLResult{}, fmt.Errorf("failed to generate unique id after %d attempts", defaultMaxRetries)
}

func (s *Shortener) AddURLs(ctx context.Context, batches []*models.ShortenBatch) ([]models.ShortenBatch, error) {
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

	result := make([]models.ShortenBatch, len(batches))
	for idx, item := range batches {
		result[idx] = *item
	}

	return result, nil
}

func (s *Shortener) GetURL(ctx context.Context, alias string) (string, error) {
	originalURL, err := s.repo.GetURLByAlias(ctx, alias)
	if err != nil {
		return "", fmt.Errorf("can't get alias: %w", err)
	}

	return originalURL, nil
}

func generateAlias() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
