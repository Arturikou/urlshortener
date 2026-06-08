package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arturikou/urlshortener/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (db *DB) AddUserURL(ctx context.Context, userID uuid.UUID, urlID int64) error {
	query := `
		INSERT INTO user_urls(user_id, url_id) 
		VALUES ($1, $2)
	`

	_, err := db.Exec(ctx, query, userID, urlID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return models.ErrURLAlreadyExists
		}
		return fmt.Errorf("failed to add user url: %w", err)
	}

	return nil
}

func (db *DB) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]models.UserUrls, error) {
	query := `
	SELECT url.url, url.alias 
	FROM user_urls 
	JOIN url ON url_id = url.id
	WHERE user_id = $1
	`

	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]models.UserUrls, 0)

	for rows.Next() {
		var url models.UserUrls
		if err = rows.Scan(&url.OriginalURL, &url.Alias); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}

	return urls, nil
}

func (db *DB) CountUsers(ctx context.Context) (int64, error) {
	q := `SELECT COUNT(DISTINCT user_id) FROM user_urls`

	var cnt int64
	err := db.QueryRow(ctx, q).Scan(&cnt)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return cnt, nil
}
