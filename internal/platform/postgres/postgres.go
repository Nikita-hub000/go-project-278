package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"links/internal/platform/postgres/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open creates a pool and verifies the connection using a generated SQL query.
func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := sqlc.New(db).CheckConnection(ctx); err != nil {
		return nil, errors.Join(fmt.Errorf("connect postgres: %w", err), db.Close())
	}
	return db, nil
}
