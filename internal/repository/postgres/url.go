package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Repo) Save(ctx context.Context, record *models.URLData) error {
	query := `INSERT INTO url (url, alias) VALUES ($1, $2)`

	_, err := s.pool.Exec(ctx, query, record.OriginalURL, record.ShortURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return models.ErrAlreadyExists
			}
		}

		return fmt.Errorf("failed to insert: %w", err)
	}

	return nil
}

func (s *Repo) Get(ctx context.Context, alias string) (string, error) {
	var originalURL string
	query := `SELECT url FROM url WHERE alias = $1`

	err := s.pool.QueryRow(ctx, query, alias).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", models.ErrNotFound
		}
		return "", fmt.Errorf("failed to get alias: %w", err)
	}

	return alias, nil
}
