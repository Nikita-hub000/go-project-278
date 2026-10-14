package visits

import (
	"context"
	"database/sql"
	"errors"
	"links/internal/platform/postgres/sqlc"
)

func GetAllVisitsHandler(ctx context.Context, db *sql.DB, pagination string, sort string, _ string) ([]Visit, error) {
	q := sqlc.New(db)

	params := sqlc.GetAllVisitsParams{PageLimit: 100, PageOffset: 0, SortRule: orderByQuery(sort)}
	if pagination != "" {
		parsed, err := parseQueryPaginationToInt(pagination)
		if err != nil {
			return nil, err
		}
		params.PageLimit = parsed.Limit
		params.PageOffset = parsed.Offset
	}

	rows, err := q.GetAllVisits(ctx, params)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var links []Visit
	for _, row := range rows {
		links = append(links, Visit{ID: row.ID, LinkID: row.LinkID, IP: row.Ip, Status: row.Status, UserAgent: row.UserAgent, CreatedAt: row.CreatedAt})
	}
	return links, nil
}

func CreateVisit(ctx context.Context, db *sql.DB, code string, ip string, userAgent string) (string, error) {
	q := sqlc.New(db)

	link, err := q.GetLinkByShortName(ctx, code)

	if err != nil {
		return "", err
	}

	_, err = q.AddVisit(ctx, sqlc.AddVisitParams{LinkID: link.ID, UserAgent: userAgent, Ip: ip, Status: 302})

	if err != nil {
		return "", err
	}

	return link.OriginalUrl, nil
}
