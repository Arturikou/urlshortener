package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsPath = "file://migrations"

type contextKey struct{ name string }

var txContextKey = &contextKey{"tx"}

type querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DB struct {
	db querier
}

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

func New(ctx context.Context, dsn string) (*DB, error) {
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

	return &DB{
		db: dbPool,
	}, nil
}

func (db *DB) getQuerier(ctx context.Context) querier {
	if tx, ok := ctx.Value(txContextKey).(pgx.Tx); ok {
		return tx
	}
	return db.db
}

func (db *DB) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txContextKey).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := db.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	err = fn(context.WithValue(ctx, txContextKey, tx))
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("rollback error: %w, original error: %w", rollbackErr, err)
		}
		return err
	}

	return tx.Commit(ctx)
}

func (db *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return db.getQuerier(ctx).Exec(ctx, sql, args...)
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.getQuerier(ctx).QueryRow(ctx, sql, args...)
}

func (db *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.getQuerier(ctx).Query(ctx, sql, args...)
}

func (db *DB) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return db.getQuerier(ctx).SendBatch(ctx, b)
}

func (db *DB) Close() {
	if pool, ok := db.db.(*pgxpool.Pool); ok {
		pool.Close()
	}
}

func (db *DB) Ping(ctx context.Context) error {
	if pool, ok := db.db.(*pgxpool.Pool); ok {
		return pool.Ping(ctx)
	}

	return nil
}
