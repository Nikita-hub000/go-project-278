package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"links/internal/links"
)

func assertShortNameConflict(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if response.Body.String() != `{"errors":{"short_name":"short name already in use"}}` {
		t.Fatalf("unexpected conflict response: %s", response.Body.String())
	}
}

func TestLinkValidationIntegration(t *testing.T) {
	db := integrationDatabase(t)
	t.Setenv("HOST_URL", "http://localhost:8080")
	router := setupRouter(db)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	create := func(name string) links.Link {
		t.Helper()
		return responseLink(t, apiRequest(t, router, http.MethodPost, "/api/links",
			`{"original_url":"https://example.com","short_name":"`+name+`","created_at":"1900-01-01T00:00:00Z"}`, http.StatusCreated))
	}
	first := create("abc")
	second := create(strings.Repeat("b", 32))
	for _, name := range []string{"", ""} {
		generated := create(name)
		if !shortNameTokenPattern.MatchString(generated.ShortName) {
			t.Fatalf("invalid generated name: %+v", generated)
		}
	}
	var createdAt time.Time
	if err := db.QueryRowContext(ctx, "SELECT created_at FROM links WHERE id = $1", first.ID).Scan(&createdAt); err != nil {
		t.Fatal(err)
	}
	if createdAt.IsZero() || createdAt.Year() == 1900 {
		t.Fatalf("created_at was not assigned by the database: %v", createdAt)
	}
	assertShortNameConflict(t, apiRequest(t, router, http.MethodPost, "/api/links",
		`{"original_url":"https://example.com/other","short_name":"abc"}`, http.StatusUnprocessableEntity))
	path := "/api/links/" + strconv.FormatInt(first.ID, 10)
	// Keeping the record's own name is allowed and updates both URL fields.
	updated := responseLink(t, apiRequest(t, router, http.MethodPut, path,
		`{"original_url":"https://example.com/updated","short_name":"abc"}`, http.StatusOK))
	if updated.ShortName != "abc" || updated.ShortURL != "http://localhost:8080/r/abc" || updated.OriginalURL != "https://example.com/updated" {
		t.Fatalf("unexpected update: %+v", updated)
	}
	assertShortNameConflict(t, apiRequest(t, router, http.MethodPut, path,
		`{"original_url":"https://example.com/rejected","short_name":"`+second.ShortName+`"}`, http.StatusUnprocessableEntity))
	unchanged := responseLink(t, apiRequest(t, router, http.MethodGet, path, "", http.StatusOK))
	if unchanged != updated {
		t.Fatalf("conflicting update changed the record: %+v, want %+v", unchanged, updated)
	}
	// Semantic and JSON errors must leave the record unchanged.
	apiRequest(t, router, http.MethodPut, path, `{"original_url":"bad","short_name":"x"}`, http.StatusUnprocessableEntity)
	apiRequest(t, router, http.MethodPut, path, `{"original_url":`, http.StatusBadRequest)
	unchanged = responseLink(t, apiRequest(t, router, http.MethodGet, path, "", http.StatusOK))
	if unchanged != updated {
		t.Fatalf("invalid request changed the record: %+v", unchanged)
	}
	// Renaming to a free name must keep short_url consistent.
	renamed := responseLink(t, apiRequest(t, router, http.MethodPut, path,
		`{"original_url":"https://example.com/renamed","short_name":"new-name"}`, http.StatusOK))
	if renamed.ShortName != "new-name" || renamed.ShortURL != "http://localhost:8080/r/new-name" {
		t.Fatalf("inconsistent renamed link: %+v", renamed)
	}
	// Both POST and PUT can omit the name; the server generates it.
	generated := responseLink(t, apiRequest(t, router, http.MethodPut, path,
		`{"original_url":"https://example.com/generated"}`, http.StatusOK))
	if !shortNameTokenPattern.MatchString(generated.ShortName) {
		t.Fatalf("PUT did not generate a name: %+v", generated)
	}
	var afterUpdate time.Time
	if err := db.QueryRowContext(ctx, "SELECT created_at FROM links WHERE id = $1", first.ID).Scan(&afterUpdate); err != nil {
		t.Fatal(err)
	}
	if !afterUpdate.Equal(createdAt) {
		t.Fatalf("created_at changed during update: %v -> %v", createdAt, afterUpdate)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM links").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("unexpected row count after rejected requests: %d", count)
	}
}

func TestConcurrentShortNameConflictIntegration(t *testing.T) {
	db := integrationDatabase(t)
	t.Setenv("HOST_URL", "http://localhost:8080")
	router := setupRouter(db)
	const requests = 8
	responses := make(chan *httptest.ResponseRecorder, requests)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range requests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(`{"original_url":"https://example.com","short_name":"same-name"}`))
			req.Header.Set("Content-Type", "application/json")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req.WithContext(ctx))
			responses <- response
		}()
	}
	close(start)
	wg.Wait()
	close(responses)
	created, conflicts := 0, 0
	for response := range responses {
		switch response.Code {
		case http.StatusCreated:
			created++
			var link links.Link
			if err := json.Unmarshal(response.Body.Bytes(), &link); err != nil {
				t.Fatal(err)
			}
			if link.ShortName != "same-name" {
				t.Fatalf("name was silently changed: %+v", link)
			}
		case http.StatusUnprocessableEntity:
			conflicts++
			assertShortNameConflict(t, response)
		default:
			t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
		}
	}
	if created != 1 || conflicts != requests-1 {
		t.Fatalf("created = %d, conflicts = %d", created, conflicts)
	}
}
