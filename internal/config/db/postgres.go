// Package db provides PostgreSQL initialization utilities
// including pool creation, connectivity checks, and migrations.
package db

import (
	"context"
	"fmt"
	"github.com/MaxDrattcev/loyalty_system_service.git/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// NewConDB creates PostgreSQL connection pool, verifies connectivity with Ping,
// runs migrations, and returns initialized pool.
func NewConDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	if cfg.Postgres.DSN == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}
	var pool *pgxpool.Pool
	pgConfig, err := pgxpool.ParseConfig(cfg.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	pgConfig.MaxConns = cfg.Postgres.MaxConns
	pgConfig.MinConns = cfg.Postgres.MinConns
	pgConfig.MaxConnLifetime = time.Duration(cfg.Postgres.MaxConnLifetime) * time.Minute
	pgConfig.MaxConnIdleTime = time.Duration(cfg.Postgres.MaxConnIdleTime) * time.Minute

	pool, pgErr := pgxpool.NewWithConfig(ctx, pgConfig)
	if pgErr != nil {
		return nil, fmt.Errorf("create pool: %w", pgErr)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pool: %w", err)
	}
	m, err := migrate.New(cfg.Postgres.PathMigration, cfg.Postgres.DSN)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		pool.Close()
		return nil, fmt.Errorf("migrate up: %w", err)
	}
	return pool, nil
}
