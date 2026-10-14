package links_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"links/internal/links"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func linkRouter(db *sql.DB) http.Handler {
	router := gin.New()
	router.POST("/api/links", links.PostLink(db))
	router.PUT("/api/links/:id", links.UpdateLink(db))
	router.GET("/api/links", links.GetAllLinks(db))
	router.GET("/api/links/:id", links.GetLinkByID(db))
	router.DELETE("/api/links/:id", links.DeleteLink(db))
	return router
}

func request(t *testing.T, router http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("%s %s: status %d, want %d; body %s", method, path, response.Code, status, response.Body.String())
	}
	return response
}

func TestRequestValidation(t *testing.T) {
	// Validating must finish before any database access, even for PUT.
	router := linkRouter(nil)
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		path := "/api/links"
		if method == http.MethodPut {
			path += "/1"
		}
		for _, tc := range []struct {
			name, body string
			fields     []string
		}{
			{"missing URL", `{}`, []string{"original_url"}},
			{"empty URL", `{"original_url":""}`, []string{"original_url"}},
			{"invalid URL", `{"original_url":"not-a-url"}`, []string{"original_url"}},
			{"short name too short", `{"original_url":"https://example.com","short_name":"ab"}`, []string{"short_name"}},
			{"short name too long", `{"original_url":"https://example.com","short_name":"` + strings.Repeat("a", 33) + `"}`, []string{"short_name"}},
			{"multiple invalid fields", `{"original_url":"bad","short_name":"ab"}`, []string{"original_url", "short_name"}},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				response := request(t, router, method, path, tc.body, http.StatusUnprocessableEntity)
				var body struct {
					Errors map[string]string `json:"errors"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if len(body.Errors) != len(tc.fields) {
					t.Fatalf("unexpected errors: %s", response.Body.String())
				}
				for _, field := range tc.fields {
					if body.Errors[field] == "" {
						t.Errorf("missing error for %s", field)
					}
				}
			})
		}
		for _, body := range []string{"", `{"original_url":`, `{"original_url":42}`, `[]`, `{"short_name":123}`, `{} {}`, `{"original_url":"https://example.com"} garbage`} {
			t.Run(method+"/invalid JSON/"+body, func(t *testing.T) {
				response := request(t, router, method, path, body, http.StatusBadRequest)
				if response.Body.String() != `{"error":"invalid request"}` {
					t.Fatalf("unexpected body: %s", response.Body.String())
				}
			})
		}
	}
}

func TestValidRequestsReachDatabase(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://localhost/unused?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	router := linkRouter(db)
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		path := "/api/links"
		if method == http.MethodPut {
			path += "/1"
		}
		for _, name := range []string{"", "abc", strings.Repeat("a", 32)} {
			t.Run(method+"/name="+name, func(t *testing.T) {
				body := `{"original_url":"https://example.com","short_name":"` + name + `"}`
				response := request(t, router, method, path, body, http.StatusInternalServerError)
				if strings.Contains(response.Body.String(), "sql:") {
					t.Fatal("internal database details exposed to the client")
				}
			})
		}
	}
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/links"}, {http.MethodGet, "/api/links/1"}, {http.MethodDelete, "/api/links/1"},
	} {
		request(t, router, tc.method, tc.path, "", http.StatusInternalServerError)
	}
}

func TestInvalidLinkID(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		for _, id := range []string{"abc", "9223372036854775808"} {
			t.Run(method+"/"+id, func(t *testing.T) {
				request(t, linkRouter(nil), method, "/api/links/"+id, "", http.StatusBadRequest)
			})
		}
	}
}
