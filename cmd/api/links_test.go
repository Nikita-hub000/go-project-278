package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
	var body struct {
		Link *links.Link `json:"link"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Link == nil {
		t.Fatalf("missing link: %s", response.Body.String())
	}
	return *body.Link
}

func assertAPIError(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if message, ok := body["error"].(string); !ok || message == "" {
		t.Fatalf("missing error message: %s", response.Body.String())
	}
	if _, ok := body["link"]; ok {
		t.Fatalf("error response contains a link: %s", response.Body.String())
	}
}

func TestLinksValidation(t *testing.T) {
	// A nil database makes accidental execution after a validation error fail.
	router := setupRouter(nil)
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		path := "/api/links"
		if method == http.MethodPut {
			path += "/1"
		}
		for _, tc := range []struct{ name, body string }{
			{"missing URL", `{"short_name":"test"}`},
			{"invalid URL", `{"original_url":"not-a-url"}`},
			{"malformed JSON", `{"original_url":`},
			{"wrong type", `{"original_url":42}`},
			{"empty body", ""},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				assertAPIError(t, apiRequest(t, router, method, path, tc.body, http.StatusBadRequest))
			})
		}
	}
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		for _, id := range []string{"abc", "9223372036854775808"} {
			t.Run(method+"/id="+id, func(t *testing.T) {
				assertAPIError(t, apiRequest(t, router, method, "/api/links/"+id, "", http.StatusBadRequest))
			})
		}
	}
}

func TestLinksDatabaseErrors(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://localhost/unused?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	// A closed pool deterministically fails without contacting PostgreSQL.
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	router := setupRouter(db)
	for _, tc := range []struct{ name, method, path, body string }{
		{"list", http.MethodGet, "/api/links", ""},
		{"get", http.MethodGet, "/api/links/1", ""},
		{"lookup before create", http.MethodPost, "/api/links", `{"original_url":"https://example.com","short_name":"test"}`},
		{"insert", http.MethodPost, "/api/links", `{"original_url":"https://example.com"}`},
		{"update", http.MethodPut, "/api/links/1", `{"original_url":"https://example.com","short_name":"test"}`},
		{"delete", http.MethodDelete, "/api/links/1", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertAPIError(t, apiRequest(t, router, tc.method, tc.path, tc.body, http.StatusInternalServerError))
		})
	}
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

func TestLinksIntegration(t *testing.T) {
	db := integrationDatabase(t)
	t.Setenv("HOST_URL", "http://localhost:8080")
	router := setupRouter(db)

	empty := apiRequest(t, router, http.MethodGet, "/api/links", "", http.StatusOK)
	var list struct {
		Links []links.Link `json:"links"`
	}
	if err := json.Unmarshal(empty.Body.Bytes(), &list); err != nil || len(list.Links) != 0 {
		t.Fatalf("expected empty list: %s (error: %v)", empty.Body.String(), err)
	}

	created := responseLink(t, apiRequest(t, router, http.MethodPost, "/api/links",
		`{"original_url":"https://example.com","short_name":"test-link"}`, http.StatusCreated))
	if created.ID <= 0 || created.OriginalURL != "https://example.com" || created.ShortName != "test-link" || created.ShortURL != "http://localhost:8080/r/test-link" {
		t.Fatalf("unexpected created link: %+v", created)
	}
	path := "/api/links/" + strconv.FormatInt(created.ID, 10)
	got := responseLink(t, apiRequest(t, router, http.MethodGet, path, "", http.StatusOK))
	if got != created {
		t.Fatalf("get = %+v, want %+v", got, created)
	}

	for i := 0; i < 2; i++ {
		automatic := responseLink(t, apiRequest(t, router, http.MethodPost, "/api/links",
			`{"original_url":"https://example.com/auto"}`, http.StatusCreated))
		if _, err := uuid.Parse(automatic.ShortName); err != nil {
			t.Fatalf("expected generated UUID, got %q", automatic.ShortName)
		}
		if automatic.ShortURL != "http://localhost:8080/r/"+automatic.ShortName {
			t.Fatalf("inconsistent generated link: %+v", automatic)
		}
	}

	duplicate := responseLink(t, apiRequest(t, router, http.MethodPost, "/api/links",
		`{"original_url":"https://example.com/duplicate","short_name":"test-link"}`, http.StatusCreated))
	if duplicate.ID == created.ID || duplicate.ShortName == created.ShortName || !strings.HasPrefix(duplicate.ShortName, "test-link") || duplicate.ShortURL != "http://localhost:8080/r/"+duplicate.ShortName {
		t.Fatalf("name collision not resolved: %+v", duplicate)
	}

	all := apiRequest(t, router, http.MethodGet, "/api/links", "", http.StatusOK)
	if err := json.Unmarshal(all.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Links) != 4 || list.Links[0] != created {
		t.Fatalf("unexpected list: %+v", list.Links)
	}
	for i := 1; i < len(list.Links); i++ {
		if list.Links[i-1].ID >= list.Links[i].ID {
			t.Fatalf("list not ordered by ID: %+v", list.Links)
		}
	}

	updated := responseLink(t, apiRequest(t, router, http.MethodPut, path,
		`{"original_url":"https://example.com/updated","short_name":"test-updated"}`, http.StatusOK))
	if updated.ID != created.ID || updated.OriginalURL != "https://example.com/updated" || updated.ShortName != "test-updated" {
		t.Fatalf("unexpected updated link: %+v", updated)
	}
	got = responseLink(t, apiRequest(t, router, http.MethodGet, path, "", http.StatusOK))
	if got != updated {
		t.Fatalf("update not persisted: got %+v, want %+v", got, updated)
	}

	apiRequest(t, router, http.MethodDelete, path, "", http.StatusNoContent)
	var count int
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM links WHERE id = $1", created.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("deleted link still exists in database")
	}
}
