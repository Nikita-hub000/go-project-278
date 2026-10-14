package visits

import "testing"

func TestOrderByQuery(t *testing.T) {
	for _, tc := range []struct{ query, want string }{
		{`["id","ASC"]`, "id ASC"},
		{`["id","DESC"]`, "id DESC"},
		{`["created_at","ASC"]`, "created_at ASC"},
		{`["created_at","DESC"]`, "created_at DESC"},
		{`["link","ASC"]`, "link_id ASC"},
		{`["link_id","DESC"]`, "link_id DESC"},
		{`["ip","ASC"]`, "ip ASC"},
		{`["ip","DESC"]`, "ip DESC"},
		{`["userAgent","ASC"]`, "user_agent ASC"},
		{`["user_agent","DESC"]`, "user_agent DESC"},
		{`["status","ASC"]`, "status ASC"},
		{`["STATUS","desc"]`, "status DESC"},
		{`["id; DROP TABLE visits","DESC"]`, "id ASC"},
		{`["id","DESC; DELETE FROM visits"]`, "id ASC"},
		{`["unknown","ASC"]`, "id ASC"},
		{`["id"]`, "id ASC"},
		{`["id",42]`, "id ASC"},
		{`null`, "id ASC"},
		{`invalid`, "id ASC"},
		{"", "id ASC"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			if got := orderByQuery(tc.query); got != tc.want {
				t.Fatalf("order = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPagination(t *testing.T) {
	for _, tc := range []struct {
		query         string
		limit, offset int32
		valid         bool
	}{
		{"[0,4]", 5, 0, true},
		{"[5,9]", 5, 5, true},
		{"[7,7]", 1, 7, true},
		{"[2147483647,2147483647]", 1, 2147483647, true},
		{"[-1,4]", 0, 0, false},
		{"[5,4]", 0, 0, false},
		{"[5,1]", 0, 0, false},
		{"[0,2147483647]", 0, 0, false},
		{"[2147483648,2147483648]", 0, 0, false},
		{"[0,1.5]", 0, 0, false},
		{"[0]", 0, 0, false},
		{"[0,4,8]", 0, 0, false},
		{"{}", 0, 0, false},
		{"null", 0, 0, false},
		{"[", 0, 0, false},
		{"", 0, 0, false},
	} {
		t.Run(tc.query, func(t *testing.T) {
			page, err := parseQueryPaginationToInt(tc.query)
			if !tc.valid {
				if err == nil {
					t.Fatalf("expected invalid range error, got %+v", page)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if page.Limit != tc.limit || page.Offset != tc.offset {
				t.Fatalf("page = %+v, want limit %d offset %d", page, tc.limit, tc.offset)
			}
		})
	}
}
