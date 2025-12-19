package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_init.sql
var initSQL string

func NewDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Connected to database")
	return pool, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, initSQL)
	return err
}
