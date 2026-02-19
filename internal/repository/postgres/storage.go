package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Repo struct {
	pool *pgxpool.Pool
}

const migrationsPath = "file://migrations"

func runMigrations(dsn string, migrationsPath string) error {
	dsn = "pgx5://" + strings.TrimPrefix(dsn, "postgres://")
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func New(ctx context.Context, dsn string) (*Repo, error) {
	if err := runMigrations(dsn, migrationsPath); err != nil {
		return nil, fmt.Errorf("migration step failed: %w", err)
	}

	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err = dbPool.Ping(pingCtx); err != nil {
		dbPool.Close()
		return nil, err
	}

	return &Repo{
		pool: dbPool,
	}, nil
}

func (r *Repo) Close() {
	r.pool.Close()
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
