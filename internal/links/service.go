package links

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"links/internal/config"

	"github.com/google/uuid"
)

func GenerateShortUrl(ctx context.Context, db *sql.DB, shortName string) (*GenerateShortURLData, error) {
	hostURL := config.Load().HostURL
	data := &GenerateShortURLData{}
	if shortName == "" {
		data.ShortName = uuid.NewString()
		data.ShortURL = hostURL + "/r/" + data.ShortName
		return data, nil
	}
	data.ShortURL = hostURL + "/r/" + shortName
	data.ShortName = shortName

	link, err := GetLinkByShortURL(ctx, db, data.ShortURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			data.ShortName = shortName
			data.ShortURL = hostURL + "/r/" + shortName
			return data, nil
		}
		return data, err
	}

	if link != nil {
		data.ShortName = shortName + uuid.NewString()
		data.ShortURL = hostURL + "/r/" + data.ShortName
		return data, nil
	}
	return data, nil
}

func parseQueryPaginationToInt(queryStr string) (*GetAllLinksPagination, error) {
	var values []int32
	if err := json.Unmarshal([]byte(queryStr), &values); err != nil {
		return nil, err
	}
	if len(values) != 2 {
		return nil, errors.New("two nums expected")
	}

	offset := values[0] - 1
	limit := values[1] - values[0] + 1

	if offset < 0 || limit < 0 {
		return nil, errors.New("invalid offset or limit")
	}
	return &GetAllLinksPagination{Limit: values[1] - values[0] + 1, Offset: values[0] - 1}, nil
}
