package pagination

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

var ErrInvalidRange = errors.New("invalid range: expected [start,end] with non-negative, ordered int32 bounds")

type Page struct{ Limit, Offset int32 }

func Parse(raw string) (*Page, error) {
	var bounds []*int64
	if err := json.Unmarshal([]byte(raw), &bounds); err != nil || len(bounds) != 2 {
		return nil, ErrInvalidRange
	}
	if bounds[0] == nil || bounds[1] == nil {
		return nil, ErrInvalidRange
	}
	start, end := *bounds[0], *bounds[1]
	if start < 0 || end < start || end > math.MaxInt32 || end-start+1 > math.MaxInt32 {
		return nil, ErrInvalidRange
	}
	return &Page{Limit: int32(end - start + 1), Offset: int32(start)}, nil
}

func ContentRange(unit string, offset int32, length int, total int64) string {
	if length == 0 {
		return fmt.Sprintf("%s */%d", unit, total)
	}
	return fmt.Sprintf("%s %d-%d/%d", unit, offset, int64(offset)+int64(length)-1, total)
}
