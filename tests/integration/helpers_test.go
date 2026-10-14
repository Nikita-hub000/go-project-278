package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"links/internal/links"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func apiRequest(t *testing.T, router http.Handler, method, path, body string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	request := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(ctx)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != wantStatus {
		t.Fatalf("%s %s: status = %d, want %d; body = %s", method, path, response.Code, wantStatus, response.Body.String())
	}
	if wantStatus == http.StatusNoContent {
		if response.Body.Len() != 0 {
			t.Fatalf("204 response must be empty: %s", response.Body.String())
		}
	} else {
		if !strings.HasPrefix(response.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("expected JSON response, got %q", response.Header().Get("Content-Type"))
		}
		if !json.Valid(response.Body.Bytes()) {
			t.Fatalf("invalid JSON response: %s", response.Body.String())
		}
	}
	return response
}

func responseLink(t *testing.T, response *httptest.ResponseRecorder) links.Link {
	t.Helper()
	var body *links.Link
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body == nil {
		t.Fatalf("missing link: %s", response.Body.String())
	}
	return *body
}

func integrationDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	admin := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { _ = admin.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	schema := "api_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatalf("create isolated test schema: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(ctx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("remove test schema: %v", err)
		}
	})
	cfg.RuntimeParams["search_path"] = schema
	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { _ = db.Close() })
	// Apply the project's current schema, never the Down section.
	migration, err := os.ReadFile("../../migrations/001_create_links.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(migration), "-- +goose Down")
	if !found {
		t.Fatal("expected a Goose Down marker in the initial migration")
	}
	if _, err := db.ExecContext(ctx, up); err != nil {
		t.Fatalf("initialize test schema: %v", err)
	}
	return db
}
