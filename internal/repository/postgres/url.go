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

func (r *Repo) InsertOrGetAlias(ctx context.Context, data models.URLData) (models.UpsertResult, error) {
	query := `WITH inserted AS (
    INSERT INTO url (url, alias)
    VALUES ($1, $2)
    ON CONFLICT (url) DO NOTHING
    RETURNING alias
	)
	SELECT alias, true FROM inserted
	UNION ALL
	SELECT alias, false FROM url WHERE url = $1
	LIMIT 1;
	`
	var upsertResult models.UpsertResult
	err := r.pool.QueryRow(ctx, query, data.OriginalURL, data.Alias).Scan(&upsertResult.Alias, &upsertResult.IsInsert)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "url_alias_key" {
			return models.UpsertResult{}, models.ErrAliasAlreadyExists
		}

		return models.UpsertResult{}, fmt.Errorf("failed to insert or get alias: %w", err)
	}

	return upsertResult, nil
}

func (r *Repo) SaveBatch(ctx context.Context, data []*models.ShortenBatch) ([]*models.ShortenBatch, error) {
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

func (r *Repo) GetURLByAlias(ctx context.Context, alias string) (string, error) {
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
