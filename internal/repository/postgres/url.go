package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/Arturikou/urlshortener/internal/service/shortener"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) Save(ctx context.Context, record models.URLData) (string, error) {
	query := `INSERT INTO url (url, alias) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, record.OriginalURL, record.Alias)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {

			if pgErr.ConstraintName == "url_url_key" {
				var actualAlias string
				errSelect := r.pool.QueryRow(ctx, `SELECT alias FROM url WHERE url = $1`, record.OriginalURL).Scan(&actualAlias)
				if errSelect != nil {
					return "", fmt.Errorf("failed to get alias: %w", errSelect)
				}
				return actualAlias, models.ErrURLAlreadyExists
			}

			if pgErr.ConstraintName == "url_alias_key" {
				return "", models.ErrAliasAlreadyExists
			}
		}
		return "", fmt.Errorf("failed to insert: %w", err)
	}

	return record.Alias, nil
}

func (r *Repo) SaveBatch(ctx context.Context, data []*shortener.ShortenBatch) ([]*shortener.ShortenBatch, error) {
	batch := &pgx.Batch{}

	for _, rec := range data {
		batch.Queue(`
            INSERT INTO url (url, alias) 
            VALUES ($1, $2) 
            ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url 
            RETURNING alias`, rec.OriginalURL, rec.Alias)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(data); i++ {
		var actualAlias string
		err := br.QueryRow().Scan(&actualAlias)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return data[i:], nil
			}

			return nil, fmt.Errorf("batch scan error at index %d: %w", i, err)
		}

		data[i].Alias = actualAlias
	}

	return nil, nil
}

func (r *Repo) Get(ctx context.Context, alias string) (string, error) {
	var originalURL string
	query := `SELECT url FROM url WHERE alias = $1`

	err := r.pool.QueryRow(ctx, query, alias).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", models.ErrNotFound
		}
		return "", fmt.Errorf("failed to get alias: %w", err)
	}

	return originalURL, nil
}
