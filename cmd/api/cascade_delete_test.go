package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"links/internal/platform/postgres/sqlc"
)

func TestDeleteLinkCascadeIntegration(t *testing.T) {
	db := integrationDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, name := range []string{"002_create_visits.sql", "003_cascade_delete_link_visits.sql"} {
		migration, err := os.ReadFile("../../migrations/" + name)
		if err != nil {
			t.Fatal(err)
		}
		up, _, found := strings.Cut(string(migration), "-- +goose Down")
		if !found {
			t.Fatalf("missing Down marker in %s", name)
		}
		if _, err := db.ExecContext(ctx, up); err != nil {
			t.Fatalf("apply %s: %v", name, err)
		}
	}
	q := sqlc.New(db)
	var ids []int64
	for _, name := range []string{"delete-me", "keep-me"} {
		link, err := q.AddLink(ctx, sqlc.AddLinkParams{
			OriginalUrl: "https://example.com/" + name,
			ShortName:   name,
			ShortUrl:    "/r/" + name,
		})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, link.ID)
		for range 2 {
			if _, err := q.AddVisit(ctx, sqlc.AddVisitParams{
				LinkID: link.ID, Ip: "192.0.2.1", UserAgent: "cascade-test", Status: 302,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}
	apiRequest(t, setupRouter(db), http.MethodDelete,
		"/api/links/"+strconv.FormatInt(ids[0], 10), "", http.StatusNoContent)
	for _, tc := range []struct {
		id   int64
		want int
	}{
		{ids[0], 0}, {ids[1], 2},
	} {
		var count int
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM visits WHERE link_id = $1", tc.id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != tc.want {
			t.Errorf("link %d: visits = %d, want %d", tc.id, count, tc.want)
		}
	}
	if _, err := q.GetLinkByID(ctx, ids[1]); err != nil {
		t.Fatalf("unrelated link was affected: %v", err)
	}
	// Verify that the Down migration restores the original foreign key behavior.
	migration, err := os.ReadFile("../../migrations/003_cascade_delete_link_visits.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, down, _ := strings.Cut(string(migration), "-- +goose Down")
	if _, err := db.ExecContext(ctx, down); err != nil {
		t.Fatal(err)
	}
	if _, err := q.DeleteLink(ctx, ids[1]); err == nil {
		t.Fatal("expected the restored foreign key to prevent deleting a link with visits")
	}
}
