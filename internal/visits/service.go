package visits

import (
	"encoding/json/v2"
	"errors"
	"strings"
)

func orderByQuery(rule string) string {
	var values []string
	if err := json.Unmarshal([]byte(rule), &values); err != nil {
		return "id ASC"
	}
	if len(values) != 2 {
		return "id ASC"
	}
	orderByMap := map[string]string{
		"id asc":          "id ASC",
		"id desc":         "id DESC",
		"created_at asc":  "created_at ASC",
		"created_at desc": "created_at DESC",
		"link asc":        "link_id ASC",
		"link desc":       "link_id DESC",
		"ip asc":          "ip ASC",
		"ip desc":         "ip DESC",
		"useragent asc":   "user_agent ASC",
		"useragent desc":  "user_agent DESC",
		"user_agent asc":  "user_agent ASC",
		"user_agent desc": "user_agent DESC",
		"link_id asc":     "link_id ASC",
		"link_id desc":    "link_id DESC",
		"status asc":      "status ASC",
		"status desc":     "status DESC",
	}

	resultQuery := strings.ToLower(values[0]) + " " + strings.ToLower(values[1])

	val, ok := orderByMap[resultQuery]
	if !ok {
		return "id ASC"
	}
	return val
}

func parseQueryPaginationToInt(queryStr string) (*GetAllVisitsPagination, error) {
	var values []int32
	if err := json.Unmarshal([]byte(queryStr), &values); err != nil {
		return nil, err
	}
	if len(values) != 2 {
		return nil, errors.New("two nums expected")
	}

	offset := values[0]
	limit := values[1] - values[0] + 1

	if offset < 0 || values[1] < values[0] || limit <= 0 {
		return nil, errors.New("invalid offset or limit")
	}
	return &GetAllVisitsPagination{Limit: values[1] - values[0] + 1, Offset: values[0]}, nil
}
