// package shortener contains the business logic for URL shortening.
package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Arturikou/urlshortener/internal/managers/audit"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/transactor"
	"github.com/Arturikou/urlshortener/internal/userctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const defaultMaxRetries = 10

//go:generate mockery
type URLRepo interface {
	AddURL(ctx context.Context, data models.URLData) (int64, error)
	GetByURL(ctx context.Context, url string) (models.URLRecord, error)
	GetByAlias(ctx context.Context, alias string) (models.URLRecord, error)
	SaveBatch(ctx context.Context, data []models.ShortenBatch) ([]models.ShortenBatch, error)
	DeleteURLs(ctx context.Context, userID uuid.UUID, aliases []string) error
	CountUrls(ctx context.Context) (int64, error)
}

//go:generate mockery
type UserURLRepo interface {
	AddUserURL(ctx context.Context, userID uuid.UUID, urlID int64) error
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error)
	CountUsers(ctx context.Context) (int64, error)
}

type AuditManager interface {
	NotifyAll(ctx context.Context, event audit.Event)
}

type Shortener struct {
	urlRepo      URLRepo
	userURLRepo  UserURLRepo
	transactor   transactor.Transactor
	auditManager AuditManager
	logger       *zap.SugaredLogger
}

func New(
	urlRepo URLRepo,
	userURLRepo UserURLRepo,
	transactor transactor.Transactor,
	auditManager AuditManager,
	logger *zap.SugaredLogger,
) *Shortener {
	return &Shortener{
		urlRepo:      urlRepo,
		userURLRepo:  userURLRepo,
		transactor:   transactor,
		auditManager: auditManager,
		logger:       logger,
	}
}

type AddURLResult struct {
	Alias    string
	IsInsert bool
}

// AddURL attempts to create a shortened alias for the provided original URL
func (s *Shortener) AddURL(ctx context.Context, originalURL string) (AddURLResult, error) {
	userID, err := userctx.UserID(ctx)
	if err != nil {
		return AddURLResult{}, fmt.Errorf("userID not found in context: %w", err)
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
			s.auditManager.NotifyAll(ctx, audit.Event{
				Timestamp: time.Now().Unix(),
				Action:    "shorten",
				UserID:    userID,
				URL:       originalURL,
			})

			return addURLResult, nil
		}

		if errors.Is(err, models.ErrAliasAlreadyExists) {
			continue
		}

		return AddURLResult{}, fmt.Errorf("can't save alias: %w", err)
	}

	return AddURLResult{}, fmt.Errorf("failed to generate unique id after %d attempts", defaultMaxRetries)
}

// AddURLs processes a batch of URLs for shortening, assigning aliases, and saving them in the repository in batches.
func (s *Shortener) AddURLs(ctx context.Context, batches []models.ShortenBatch) ([]models.ShortenBatch, error) {
	batchSize := 1000

	for i := 0; i < len(batches); i += batchSize {
		end := i + batchSize
		if end > len(batches) {
			end = len(batches)
		}

		sub := batches[i:end]
		for j := range sub {
			alias, err := generateAlias()
			if err != nil {
				return nil, fmt.Errorf("failed to generate alias: %w", err)
			}
			sub[j].Alias = alias
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

	return batches, nil
}

// GetURL retrieves the original URL associated with the given alias from the repository and audits the access event.
func (s *Shortener) GetURL(ctx context.Context, alias string) (string, error) {
	urlRecord, err := s.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		return "", fmt.Errorf("can't get alias: %w", err)
	}

	userID, err := userctx.UserID(ctx)
	if err != nil {
		return "", fmt.Errorf("userID not found in context: %w", err)
	}

	s.auditManager.NotifyAll(ctx, audit.Event{
		Timestamp: time.Now().Unix(),
		Action:    "follow",
		UserID:    userID,
		URL:       urlRecord.URL,
	})

	return urlRecord.URL, nil
}

// GetUserURLs retrieves all shortened URLs associated with a specific user ID from the repository.
func (s *Shortener) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error) {
	return s.userURLRepo.GetUserURLs(ctx, userID)
}

// addUserURL associates a user with a shortened URL, creating the URL in the repository if it does not already exist.
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

// DeleteUserURLs removes a list of URLs associated with a user from the repository based on their aliases.
func (s *Shortener) DeleteUserURLs(ctx context.Context, userID uuid.UUID, aliases []string) error {
	if err := s.urlRepo.DeleteURLs(ctx, userID, aliases); err != nil {
		return fmt.Errorf("failed to delete urls: %w", err)
	}

	return nil
}

// generateAlias generates a random URL-safe string of 16 bytes, encodes it using base64, and returns it.
func generateAlias() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
