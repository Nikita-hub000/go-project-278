package links

import (
	"context"
	"database/sql"
	"errors"
	"links/internal/platform/postgres/sqlc"
)

func GetLinkByIDHandler(ctx context.Context, db *sql.DB, id int64) (*Link, error) {
	q := sqlc.New(db)

	row, err := q.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &Link{
		ID:          row.ID,
		ShortName:   row.ShortName,
		ShortURL:    row.ShortUrl,
		OriginalURL: row.OriginalUrl,
	}, nil
}

func GetLinkByShortURL(ctx context.Context, db *sql.DB, shortURL string) (*Link, error) {
	q := sqlc.New(db)

	link, err := q.GetLinkByShortUrl(ctx, shortURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &Link{ID: link.ID, ShortURL: link.ShortUrl, ShortName: link.ShortName, OriginalURL: link.OriginalUrl}, nil
}

func GetAllLinksHandler(ctx context.Context, db *sql.DB, pagination string, sort string, _ string) ([]Link, error) {
	q := sqlc.New(db)

	params := sqlc.GetAllLinksParams{PageLimit: 100, PageOffset: 0, SortRule: orderByQuery(sort)}
	if pagination != "" {
		parsed, err := parseQueryPaginationToInt(pagination)
		if err != nil {
			return nil, err
		}
		params.PageLimit = parsed.Limit
		params.PageOffset = parsed.Offset
	}

	rows, err := q.GetAllLinks(ctx, params)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var links []Link
	for _, row := range rows {
		links = append(links, Link{ID: row.ID, ShortName: row.ShortName, ShortURL: row.ShortUrl, OriginalURL: row.OriginalUrl})
	}
	return links, nil
}

func CreateLink(ctx context.Context, db *sql.DB, link *Link) (*Link, error) {
	q := sqlc.New(db)

	row, err := q.AddLink(ctx, sqlc.AddLinkParams{ShortUrl: link.ShortURL, OriginalUrl: link.OriginalURL, ShortName: link.ShortName})

	if err != nil {
		return nil, err
	}

	return &Link{
		ID:          row.ID,
		ShortName:   row.ShortName,
		ShortURL:    row.ShortUrl,
		OriginalURL: row.OriginalUrl,
	}, nil
}

func UpdateLinkHandler(ctx context.Context, db *sql.DB, params sqlc.UpdateLinkParams) (*Link, error) {
	q := sqlc.New(db)
	link, err := q.UpdateLink(ctx, sqlc.UpdateLinkParams{ID: params.ID, ShortName: params.ShortName, OriginalUrl: params.OriginalUrl, ShortUrl: params.ShortUrl})

	if err != nil {
		return nil, err
	}

	return &Link{
		ID:          link.ID,
		ShortName:   link.ShortName,
		ShortURL:    link.ShortUrl,
		OriginalURL: link.OriginalUrl,
	}, nil
}

func DeleteLinkByIDHandler(ctx context.Context, db *sql.DB, id int64) error {
	q := sqlc.New(db)
	affected, err := q.DeleteLink(ctx, id)

	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no such link")
	}

	return nil
}
