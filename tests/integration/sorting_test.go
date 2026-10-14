package integration

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"links/internal/links"
	"links/internal/platform/postgres/sqlc"
	"links/internal/visits"
)

func TestListSortingIntegration(t *testing.T) {
	db := integrationDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	migration, err := os.ReadFile("../../migrations/002_create_visits.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(migration), "-- +goose Down")
	if !found {
		t.Fatal("missing Down marker in visits migration")
	}
	if _, err := db.ExecContext(ctx, up); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO links (original_url, short_name, short_url, created_at) VALUES
		('https://example.com/1', 'one', '/r/one', '2026-01-03'),
		('https://example.com/2', 'two', '/r/two', '2026-01-01'),
		('https://example.com/3', 'three', '/r/three', '2026-01-02');
		INSERT INTO visits (link_id, ip, user_agent, status, created_at) VALUES
		(3, '192.0.2.3', 'C', 404, '2026-01-03'),
		(1, '192.0.2.1', 'A', 200, '2026-01-01'),
		(2, '192.0.2.2', 'B', 301, '2026-01-02');
	`); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		sort string
		want []int64
	}{
		{`["id","ASC"]`, []int64{1, 2, 3}},
		{`["id","DESC"]`, []int64{3, 2, 1}},
		{`["created_at","ASC"]`, []int64{2, 3, 1}},
		{`["created_at","DESC"]`, []int64{1, 3, 2}},
		{`["id; DROP TABLE links","DESC"]`, []int64{1, 2, 3}},
		{`invalid`, []int64{1, 2, 3}},
	} {
		t.Run("links/"+tc.sort, func(t *testing.T) {
			rows, err := links.GetAllLinksHandler(ctx, db, "", tc.sort, "invalid filter is ignored")
			if err != nil {
				t.Fatal(err)
			}
			var ids []int64
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
			if !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("IDs = %v, want %v", ids, tc.want)
			}
		})
	}
	for _, field := range []string{"id", "created_at", "link_id", "link", "ip", "user_agent", "userAgent", "status"} {
		for _, direction := range []string{"ASC", "DESC"} {
			t.Run("visits/"+field+"/"+direction, func(t *testing.T) {
				sort := `["` + field + `","` + direction + `"]`
				rows, err := visits.GetAllVisitsHandler(ctx, db, "", sort, "{}")
				if err != nil {
					t.Fatal(err)
				}
				want := []int64{2, 3, 1}
				if field == "id" {
					want = []int64{1, 2, 3}
				}
				if direction == "DESC" {
					want[0], want[2] = want[2], want[0]
				}
				var ids []int64
				for _, row := range rows {
					ids = append(ids, row.ID)
				}
				if !reflect.DeepEqual(ids, want) {
					t.Fatalf("IDs = %v, want %v", ids, want)
				}
			})
		}
	}
	// LIMIT and OFFSET must be applied after sorting.
	q := sqlc.New(db)
	page, err := q.GetAllLinks(ctx, sqlc.GetAllLinksParams{SortRule: "created_at DESC", PageLimit: 1, PageOffset: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 1 || page[0].ID != 3 {
		t.Fatalf("unexpected sorted page: %+v", page)
	}
}
