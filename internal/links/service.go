package links

import (
	"context"
	"database/sql"
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
