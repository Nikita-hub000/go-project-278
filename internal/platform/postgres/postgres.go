package postgres

import (
	"context"
	"errors"
	"fmt"

	"links/internal/platform/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open creates a pool and verifies the connection using a generated SQL query.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := sqlc.New(pool).CheckConnection(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return pool, nil
}
