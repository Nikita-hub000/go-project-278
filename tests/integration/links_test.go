package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"

	"links/internal/links"
)

var shortNameTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8}$`)

func TestLinksIntegration(t *testing.T) {
	db := integrationDatabase(t)
	t.Setenv("HOST_URL", "http://localhost:8080")
	router := setupRouter(db)

	empty := apiRequest(t, router, http.MethodGet, "/api/links", "", http.StatusOK)
	var list []links.Link
	if err := json.Unmarshal(empty.Body.Bytes(), &list); err != nil || list == nil || len(list) != 0 {
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
		if !shortNameTokenPattern.MatchString(automatic.ShortName) {
			t.Fatalf("expected an 8-character URL-safe generated name, got %q", automatic.ShortName)
		}
		if automatic.ShortURL != "http://localhost:8080/r/"+automatic.ShortName {
			t.Fatalf("inconsistent generated link: %+v", automatic)
		}
	}

	conflict := apiRequest(t, router, http.MethodPost, "/api/links",
		`{"original_url":"https://example.com/duplicate","short_name":"test-link"}`, http.StatusUnprocessableEntity)
	if conflict.Body.String() != `{"errors":{"short_name":"short name already in use"}}` {
		t.Fatalf("unexpected conflict response: %s", conflict.Body.String())
	}

	all := apiRequest(t, router, http.MethodGet, "/api/links", "", http.StatusOK)
	if err := json.Unmarshal(all.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 || list[0] != created {
		t.Fatalf("unexpected list: %+v", list)
	}
	for i := 1; i < len(list); i++ {
		if list[i-1].ID >= list[i].ID {
			t.Fatalf("list not ordered by ID: %+v", list)
		}
	}
	for _, tc := range []struct {
		name, query string
		start, end  int
	}{
		{"first", "%5B0%2C0%5D", 0, 1},
		{"middle", "%5B1%2C2%5D", 1, 3},
		{"past end", "%5B5%2C10%5D", 3, 3},
	} {
		t.Run("pagination/"+tc.name, func(t *testing.T) {
			response := apiRequest(t, router, http.MethodGet, "/api/links?range="+tc.query, "", http.StatusOK)
			var page []links.Link
			if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
				t.Fatal(err)
			}
			want := list[tc.start:tc.end]
			if len(page) != len(want) {
				t.Fatalf("page length = %d, want %d", len(page), len(want))
			}
			for i := range want {
				if page[i] != want[i] {
					t.Errorf("page[%d] = %+v, want %+v", i, page[i], want[i])
				}
			}
		})
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
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		response := apiRequest(t, router, method, path,
			`{"original_url":"https://example.com/deleted","short_name":"deleted"}`, http.StatusNotFound)
		if response.Body.String() != `{"error":"link not found"}` {
			t.Fatalf("unexpected response for deleted link: %s", response.Body.String())
		}
	}
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

func TestMissingLinkIntegration(t *testing.T) {
	db := integrationDatabase(t)
	router := setupRouter(db)
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			response := apiRequest(t, router, method, "/api/links/999999",
				`{"original_url":"https://example.com","short_name":"missing"}`, http.StatusNotFound)
			if response.Body.String() != `{"error":"link not found"}` {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
		})
	}
}
