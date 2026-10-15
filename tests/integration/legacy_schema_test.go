package integration

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"links/internal/platform/postgres/sqlc"
)

func TestRestoreShortURLMigrationIntegration(t *testing.T) {
	for _, legacy := range []bool{true, false} {
		name := "current schema"
		if legacy {
			name = "legacy schema"
		}
		t.Run(name, func(t *testing.T) {
			db := integrationDatabase(t)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if legacy {
				// Reproduce the first deployed version of migration 001.
				if _, err := db.ExecContext(ctx, "ALTER TABLE links DROP COLUMN short_url"); err != nil {
					t.Fatal(err)
				}
				if _, err := db.ExecContext(ctx, `INSERT INTO links (original_url, short_name) VALUES ('https://example.com/old', 'old-link')`); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, err := db.ExecContext(ctx, `INSERT INTO links (original_url, short_name, short_url) VALUES ('https://example.com/old', 'old-link', 'https://short.example/r/old-link')`); err != nil {
					t.Fatal(err)
				}
			}
			migration, err := os.ReadFile("../../db/migrations/004_restore_links_short_url.sql")
			if err != nil {
				t.Fatal(err)
			}
			up, down, found := strings.Cut(string(migration), "-- +goose Down")
			if !found {
				t.Fatal("missing Down marker")
			}
			// Applying the repair twice must preserve existing URLs and rows.
			for range 2 {
				if _, err := db.ExecContext(ctx, up); err != nil {
					t.Fatal(err)
				}
			}
			row, err := sqlc.New(db).GetLinkByShortName(ctx, "old-link")
			if err != nil {
				t.Fatal(err)
			}
			wantURL := "https://short.example/r/old-link"
			if legacy {
				wantURL = "/r/old-link"
			}
			if row.ShortUrl != wantURL || row.OriginalUrl != "https://example.com/old" {
				t.Fatalf("existing link was corrupted: %+v", row)
			}
			apiRequest(t, setupRouter(db), http.MethodGet,
				"/api/links?filter=%7B%7D&range=%5B0%2C4%5D&sort=%5B%22id%22%2C%22ASC%22%5D", "", http.StatusOK)
			// The repaired column still enforces the original constraints.
			if _, err := db.ExecContext(ctx, `INSERT INTO links (original_url, short_name, short_url) VALUES ('https://example.com/new', 'new-link', $1)`, wantURL); err == nil {
				t.Error("duplicate short_url was accepted")
			}
			if _, err := db.ExecContext(ctx, `INSERT INTO links (original_url, short_name) VALUES ('https://example.com/new', 'new-link')`); err == nil {
				t.Error("NULL short_url was accepted")
			}
			if _, err := db.ExecContext(ctx, down); err != nil {
				t.Fatal(err)
			}
			if _, err := sqlc.New(db).GetLinkByShortName(ctx, "old-link"); err != nil {
				t.Fatalf("rollback removed baseline data: %v", err)
			}
		})
	}
}
