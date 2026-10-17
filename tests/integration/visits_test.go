package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"links/internal/platform/postgres/sqlc"
	"links/internal/visits"
)

func visitsDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db := integrationDatabase(t)
	migration, err := os.ReadFile("../../db/migrations/002_create_visits.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(migration), "-- +goose Down")
	if !found {
		t.Fatal("missing Down marker in visits migration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, up); err != nil {
		t.Fatal(err)
	}
	return db
}

func addVisitedLink(t *testing.T, ctx context.Context, db *sql.DB, name string) sqlc.Link {
	t.Helper()
	link, err := sqlc.New(db).AddLink(ctx, sqlc.AddLinkParams{
		ShortName: name, ShortUrl: "/r/" + name, OriginalUrl: "https://example.com/" + name,
	})
	if err != nil {
		t.Fatal(err)
	}
	return link
}

func redirectRequest(t *testing.T, router http.Handler, code, ip, agent string, status int) *httptest.ResponseRecorder {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/r/"+code, nil).WithContext(ctx)
	req.RemoteAddr = ip + ":12345"
	req.Header.Set("User-Agent", agent)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("redirect status = %d, want %d; body %s", response.Code, status, response.Body.String())
	}
	return response
}

func decodeVisits(t *testing.T, response *httptest.ResponseRecorder) []visits.Visit {
	t.Helper()
	var list []visits.Visit
	if err := json.Unmarshal(response.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if list == nil {
		t.Fatalf("expected a JSON array, got %s", response.Body.String())
	}
	return list
}

func TestVisitsRedirectIntegration(t *testing.T) {
	db := visitsDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	router := setupRouter(db)
	first := addVisitedLink(t, ctx, db, "first")
	second := addVisitedLink(t, ctx, db, "second")
	// An empty list is [] so the UI can render it without a special null case.
	empty := decodeVisits(t, apiRequest(t, router, http.MethodGet, "/api/link_visits", "", http.StatusOK))
	if len(empty) != 0 {
		t.Fatalf("expected no visits, got %+v", empty)
	}
	before := time.Now().Add(-time.Second)
	for _, link := range []sqlc.Link{first, first, second} {
		response := redirectRequest(t, router, link.ShortName, "203.0.113.7", "visits-test/1.0", http.StatusFound)
		if got := response.Header().Get("Location"); got != link.OriginalUrl {
			t.Errorf("Location = %q, want %q", got, link.OriginalUrl)
		}
		if got := response.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, want no-store", got)
		}
	}
	list := decodeVisits(t, apiRequest(t, router, http.MethodGet, "/api/link_visits", "", http.StatusOK))
	if len(list) != 3 {
		t.Fatalf("expected all three visits, got %+v", list)
	}
	for i, visit := range list {
		wantLinkID := first.ID
		if i == 2 {
			wantLinkID = second.ID
		}
		if visit.LinkID != wantLinkID || visit.IP != "203.0.113.7" || visit.UserAgent != "visits-test/1.0" || visit.Status != http.StatusFound {
			t.Errorf("incorrect visit data: %+v", visit)
		}
		if visit.CreatedAt.Before(before) || visit.CreatedAt.After(time.Now().Add(time.Second)) {
			t.Errorf("invalid created_at: %v", visit.CreatedAt)
		}
		if visit.ID <= 0 || (i > 0 && visit.ID <= list[i-1].ID) {
			t.Errorf("visit ID was not generated independently: %+v", visit)
		}
	}
	for _, tc := range []struct {
		name, query string
		want        []int64
	}{
		{"first page", "range=%5B0%2C1%5D", []int64{list[0].ID, list[1].ID}},
		{"one item", "range=%5B2%2C2%5D", []int64{list[2].ID}},
		{"beyond end", "range=%5B3%2C9%5D", nil},
		{"descending page", "range=%5B0%2C1%5D&sort=%5B%22id%22%2C%22DESC%22%5D", []int64{list[2].ID, list[1].ID}},
		{"ignored filter", "filter=" + url.QueryEscape(`{"link_id":999999}`), []int64{list[0].ID, list[1].ID, list[2].ID}},
		{"malformed filter", "filter=not-json", []int64{list[0].ID, list[1].ID, list[2].ID}},
		{"invalid sort fallback", "sort=invalid", []int64{list[0].ID, list[1].ID, list[2].ID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := decodeVisits(t, apiRequest(t, router, http.MethodGet, "/api/link_visits?"+tc.query, "", http.StatusOK))
			if len(page) != len(tc.want) {
				t.Fatalf("page = %+v, want IDs %v", page, tc.want)
			}
			for i, id := range tc.want {
				if page[i].ID != id {
					t.Errorf("page[%d].ID = %d, want %d", i, page[i].ID, id)
				}
			}
		})
	}
}

func TestVisitsDefaultLimitIntegration(t *testing.T) {
	db := visitsDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	link := addVisitedLink(t, ctx, db, "many-visits")
	if _, err := db.ExecContext(ctx, `INSERT INTO visits (link_id, ip, user_agent, status)
		SELECT $1, '203.0.113.7', 'limit-test', 302 FROM generate_series(1,101)`, link.ID); err != nil {
		t.Fatal(err)
	}
	router := setupRouter(db)
	page := decodeVisits(t, apiRequest(t, router, http.MethodGet, "/api/link_visits", "", http.StatusOK))
	if len(page) != 100 {
		t.Fatalf("default list length = %d, want 100", len(page))
	}
	last := decodeVisits(t, apiRequest(t, router, http.MethodGet, "/api/link_visits?range=%5B100%2C100%5D", "", http.StatusOK))
	if len(last) != 1 || last[0].ID <= page[99].ID {
		t.Fatalf("unexpected last page: %+v", last)
	}
}

func TestVisitCreationErrorsIntegration(t *testing.T) {
	db := visitsDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	router := setupRouter(db)
	link := addVisitedLink(t, ctx, db, "reject-insert")
	if _, err := db.ExecContext(ctx, `ALTER TABLE visits ADD CONSTRAINT reject_test_agent CHECK (user_agent <> 'reject-visit')`); err != nil {
		t.Fatal(err)
	}
	missing := redirectRequest(t, router, "missing-code", "203.0.113.7", "normal-agent", http.StatusNotFound)
	if missing.Header().Get("Location") != "" || missing.Body.String() != `{"error":"link not found"}` {
		t.Fatalf("unexpected missing link response: %s", missing.Body.String())
	}
	var diagnostics bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&diagnostics, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })
	response := redirectRequest(t, router, link.ShortName, "203.0.113.7", "reject-visit", http.StatusFound)
	if response.Header().Get("Location") != link.OriginalUrl || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("analytics failure prevented redirect: %v", response.Header())
	}
	for _, field := range []string{`"operation":"create_visit"`, `"code":"reject-insert"`, `"error":`, "reject_test_agent"} {
		if !strings.Contains(diagnostics.String(), field) {
			t.Errorf("missing diagnostic %s: %s", field, diagnostics.String())
		}
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM visits").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("failed requests left %d visits", count)
	}
}
