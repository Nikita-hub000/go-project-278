package pagination

import "testing"

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		raw           string
		offset, limit int32
	}{
		{"[0,9]", 0, 10}, {"[5,5]", 5, 1}, {"[2147483647,2147483647]", 2147483647, 1},
	} {
		page, err := Parse(tc.raw)
		if err != nil || page.Offset != tc.offset || page.Limit != tc.limit {
			t.Fatalf("Parse(%s) = %v, %v", tc.raw, page, err)
		}
	}
	for _, raw := range []string{"oops", "[]", "[0]", "[-1,1]", "[5,4]", "[0,2147483647]", "[2147483648,2147483648]", "[0,9223372036854775808]", "[null,1]", "[0,1.5]"} {
		if _, err := Parse(raw); err != ErrInvalidRange {
			t.Errorf("Parse(%s): %v", raw, err)
		}
	}
}

func TestContentRange(t *testing.T) {
	for _, tc := range []struct {
		offset int32
		length int
		total  int64
		want   string
	}{
		{0, 1, 1, "links 0-0/1"}, {4, 1, 5, "links 4-4/5"}, {5, 0, 5, "links */5"}, {0, 0, 0, "links */0"},
	} {
		if got := ContentRange("links", tc.offset, tc.length, tc.total); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}
