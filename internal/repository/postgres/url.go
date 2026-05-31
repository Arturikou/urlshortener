package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (db *DB) AddURL(ctx context.Context, data models.URLData) (int64, error) {
	var id int64
	query := `INSERT INTO url (url, alias) VALUES ($1, $2) RETURNING id`

	err := db.QueryRow(ctx, query, data.OriginalURL, data.Alias).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.ConstraintName {
			case "url_url_key":
				return 0, models.ErrURLAlreadyShorted
			case "url_alias_key":
				return 0, models.ErrAliasAlreadyExists
			}
		}

		return 0, fmt.Errorf("insert failed: %w", err)
	}

	return id, nil
}

func (db *DB) SaveBatch(ctx context.Context, data []models.ShortenBatch) ([]models.ShortenBatch, error) {
	batch := &pgx.Batch{}

	for _, rec := range data {
		batch.Queue(`
            INSERT INTO url (url, alias) 
            VALUES ($1, $2) 
            ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url 
            RETURNING alias`, rec.OriginalURL, rec.Alias)
	}

	br := db.SendBatch(ctx, batch)
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

func (db *DB) GetByAlias(ctx context.Context, alias string) (models.URLRecord, error) {
	var record models.URLRecord
	var isDeleted bool
	query := `SELECT id, alias, url, is_deleted FROM url WHERE alias = $1`

	err := db.QueryRow(ctx, query, alias).Scan(&record.ID, &record.Alias, &record.URL, &isDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.URLRecord{}, models.ErrNotFound
		}
		return models.URLRecord{}, fmt.Errorf("failed to get alias: %w", err)
	}

	if isDeleted {
		return models.URLRecord{}, models.ErrURLDeleted
	}

	return record, nil
}

func (db *DB) GetByURL(ctx context.Context, url string) (models.URLRecord, error) {
	var record models.URLRecord
	var isDeleted bool

	query := `SELECT id, alias, url, is_deleted FROM url WHERE url = $1`
	err := db.QueryRow(ctx, query, url).Scan(&record.ID, &record.Alias, &record.URL, &isDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.URLRecord{}, models.ErrNotFound
		}
		return models.URLRecord{}, fmt.Errorf("failed to get alias: %w", err)
	}

	if isDeleted {
		return models.URLRecord{}, models.ErrURLDeleted
	}

	return record, nil
}

func (db *DB) DeleteURLs(ctx context.Context, userID uuid.UUID, aliases []string) error {
	if len(aliases) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, alias := range aliases {
		batch.Queue(`
		UPDATE url
		SET is_deleted = TRUE
		FROM user_urls
		WHERE url.id = user_urls.url_id
		  AND user_urls.user_id = $1
		  AND url.alias = $2
	`, userID, alias)
	}

	br := db.SendBatch(ctx, batch)
	defer br.Close()

	for range aliases {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}
	}

	return nil
}

func (db *DB) CountUrls(ctx context.Context) (int64, error) {
	q := `SELECT COUNT(url) FROM url`

	var cnt int64
	err := db.QueryRow(ctx, q).Scan(&cnt)
	if err != nil {
		return 0, fmt.Errorf("failed to count urls: %w", err)
	}

	return cnt, nil
}
