package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/Arturikou/urlshortener/internal/middleware"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/transactor"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const defaultMaxRetries = 10

//go:generate mockery
type URLRepo interface {
	AddURL(ctx context.Context, data models.URLData) (int64, error)
	GetByURL(ctx context.Context, url string) (models.URLRecord, error)
	GetByAlias(ctx context.Context, alias string) (models.URLRecord, error)
	SaveBatch(ctx context.Context, data []*models.ShortenBatch) ([]*models.ShortenBatch, error)
	DeleteURLs(ctx context.Context, userID uuid.UUID, aliases []string) error
}

//go:generate mockery
type UserURLRepo interface {
	AddUserURL(ctx context.Context, userID uuid.UUID, urlID int64) error
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error)
}

type Shortener struct {
	urlRepo     URLRepo
	userURLRepo UserURLRepo
	transactor  transactor.Transactor
	logger      *zap.SugaredLogger
}

func New(
	urlRepo URLRepo,
	userURLRepo UserURLRepo,
	transactor transactor.Transactor,
	logger *zap.SugaredLogger,
) *Shortener {
	return &Shortener{
		urlRepo:     urlRepo,
		userURLRepo: userURLRepo,
		transactor:  transactor,
		logger:      logger,
	}
}

type AddURLResult struct {
	Alias    string
	IsInsert bool
}

func (s *Shortener) AddURL(ctx context.Context, originalURL string) (AddURLResult, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return AddURLResult{}, fmt.Errorf("userID not found in context")
	}

	for i := 0; i < defaultMaxRetries; i++ {
		alias, err := generateAlias()
		if err != nil {
			return AddURLResult{}, fmt.Errorf("failed to generate alias: %w", err)
		}

		urlData := models.URLData{
			OriginalURL: originalURL,
			Alias:       alias,
		}

		addURLResult, err := s.addUserURL(ctx, userID, urlData)
		if err == nil {
			return addURLResult, nil
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
			remaining, err := s.urlRepo.SaveBatch(ctx, currentBatch)
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
	urlRecord, err := s.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		return "", fmt.Errorf("can't get alias: %w", err)
	}

	if urlRecord.DeletedFlag {
		return "", models.ErrURLDeleted
	}

	return urlRecord.URL, nil
}

func (s *Shortener) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error) {
	return s.userURLRepo.GetUserURLs(ctx, userID)
}

func (s *Shortener) addUserURL(ctx context.Context, userID uuid.UUID, urlData models.URLData) (AddURLResult, error) {
	err := s.transactor.Transaction(ctx, func(ctx context.Context) error {
		urlID, err := s.urlRepo.AddURL(ctx, urlData)
		if err != nil {
			return err
		}

		return s.userURLRepo.AddUserURL(ctx, userID, urlID)
	})

	if err == nil {
		return AddURLResult{Alias: urlData.Alias, IsInsert: true}, nil
	}

	if errors.Is(err, models.ErrURLAlreadyShorted) {
		urlRecord, err := s.urlRepo.GetByURL(ctx, urlData.OriginalURL)
		if err != nil {
			return AddURLResult{}, err
		}

		err = s.userURLRepo.AddUserURL(ctx, userID, urlRecord.ID)
		if err == nil || errors.Is(err, models.ErrURLAlreadyExists) {
			return AddURLResult{Alias: urlRecord.Alias, IsInsert: false}, nil
		}

		return AddURLResult{}, fmt.Errorf("can't add user url: %w", err)
	}

	return AddURLResult{}, err
}

func (s *Shortener) DeleteUserURLs(ctx context.Context, aliases []string, userID uuid.UUID) error {
	const batchSize = 100

	for i := 0; i < len(aliases); i += batchSize {
		end := i + batchSize
		if end > len(aliases) {
			end = len(aliases)
		}

		batch := aliases[i:end]
		if err := s.urlRepo.DeleteURLs(ctx, userID, batch); err != nil {
			s.logger.Errorf("failed to delete urls: %v", err)
		}
	}

	return nil
}

func generateAlias() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
