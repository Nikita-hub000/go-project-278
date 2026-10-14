package visits_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"links/internal/visits"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func visitsRouter(db *sql.DB) http.Handler {
	router := gin.New()
	router.GET("/api/link_visits", visits.GetAllVisits(db))
	router.GET("/r/:code", visits.RedirectToShortLink(db))
	return router
}

func assertVisitError(t *testing.T, router http.Handler, path, wantBody string) {
	t.Helper()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	if response.Code != http.StatusInternalServerError || response.Body.String() != wantBody {
		t.Fatalf("GET %s: status %d, body %s", path, response.Code, response.Body.String())
	}
	if got := response.Header().Get("Location"); got != "" {
		t.Fatalf("error response must not redirect: %q", got)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
}

func TestDatabaseErrors(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://localhost/unused?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	router := visitsRouter(db)
	t.Run("list", func(t *testing.T) {
		assertVisitError(t, router, "/api/link_visits", `{"error":"failed to retrieve visits"}`)
	})
	t.Run("redirect", func(t *testing.T) {
		assertVisitError(t, router, "/r/existing-code", `{"error":"failed to create visit"}`)
	})
}

func TestInvalidRangeDoesNotAccessDatabase(t *testing.T) {
	for _, query := range []string{"invalid", "[0]", "[-1,4]", "[5,4]", "[0,2147483647]"} {
		t.Run(query, func(t *testing.T) {
			// No recovery middleware or DB: any accidental query fails the test.
			assertVisitError(t, visitsRouter(nil), "/api/link_visits?range="+url.QueryEscape(query), `{"error":"failed to retrieve visits"}`)
		})
	}
}
