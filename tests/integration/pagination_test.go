package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestContentRangeIntegration(t *testing.T) {
	db := visitsDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	router := setupRouter(db)
	for _, resource := range []string{"links", "link_visits"} {
		response := apiRequest(t, router, http.MethodGet, "/api/"+resource+"?range="+url.QueryEscape("[0,9]"), "", http.StatusOK)
		if got := response.Header().Get("Content-Range"); got != resource+" */0" {
			t.Fatalf("empty collection: %q", got)
		}
	}
	for i := 0; i < 5; i++ {
		link := addVisitedLink(t, ctx, db, "page-"+strconv.Itoa(i))
		if _, err := db.ExecContext(ctx, "INSERT INTO visits (link_id,ip,user_agent,status) VALUES ($1,'127.0.0.1','pagination-test',302)", link.ID); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			for _, resource := range []string{"links", "link_visits"} {
				response := apiRequest(t, router, http.MethodGet, "/api/"+resource+"?range="+url.QueryEscape("[0,9]"), "", http.StatusOK)
				if got := response.Header().Get("Content-Range"); got != resource+" 0-0/1" {
					t.Fatalf("single item: %q", got)
				}
			}
		}
	}
	for _, resource := range []string{"links", "link_visits"} {
		for _, tc := range []struct {
			name, query, suffix string
			size                int
		}{
			{"default", "", "0-4/5", 5},
			{"first", "[0,1]", "0-1/5", 2},
			{"middle", "[2,3]", "2-3/5", 2},
			{"last", "[4,9]", "4-4/5", 1},
			{"past end", "[5,10]", "*/5", 0},
		} {
			t.Run(resource+"/"+tc.name, func(t *testing.T) {
				response := apiRequest(t, router, http.MethodGet, "/api/"+resource+"?range="+url.QueryEscape(tc.query), "", http.StatusOK)
				if got := response.Header().Get("Content-Range"); got != resource+" "+tc.suffix {
					t.Fatalf("Content-Range = %q, want %q", got, resource+" "+tc.suffix)
				}
				var rows []json.RawMessage
				if err := json.Unmarshal(response.Body.Bytes(), &rows); err != nil {
					t.Fatal(err)
				}
				if rows == nil || len(rows) != tc.size {
					t.Fatalf("unexpected page: %s", response.Body.String())
				}
			})
		}
	}
}
