package links

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"links/internal/config"
	"strings"
)

const randomShortNameBytes = 6
const maxShortNameGenerationAttempts = 5

func GenerateShortUrl(ctx context.Context, db *sql.DB, shortName string) (*GenerateShortURLData, error) {
	hostURL := config.Load().HostURL
	if shortName == "" {
		return generateUniqueShortName(ctx, db, hostURL, "")
	}
	shortURL := hostURL + "/r/" + shortName

	link, err := GetLinkByShortURL(ctx, db, shortURL)

	if err != nil {
		return nil, err
	}

	if link != nil {
		return generateUniqueShortName(ctx, db, hostURL, shortName)
	}
	return &GenerateShortURLData{ShortName: shortName, ShortURL: shortURL}, nil
}

func generateUniqueShortName(ctx context.Context, db *sql.DB, hostURL, prefix string) (*GenerateShortURLData, error) {
	for range maxShortNameGenerationAttempts {
		suffix, err := generateRandomShortName()
		if err != nil {
			return nil, err
		}
		generatedShortName := suffix
		if prefix != "" {
			generatedShortName = prefix + "-" + suffix
		}
		shortURL := hostURL + "/r/" + generatedShortName
		link, err := GetLinkByShortURL(ctx, db, shortURL)
		if err != nil {
			return nil, err
		}
		if link == nil {
			return &GenerateShortURLData{
				ShortName: generatedShortName,
				ShortURL:  shortURL,
			}, nil
		}
	}
	return nil, errors.New("could not generate a unique short name")
}

func generateRandomShortName() (string, error) {
	bytes := make([]byte, randomShortNameBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func parseQueryPaginationToInt(queryStr string) (*GetAllLinksPagination, error) {
	var values []int32
	if err := json.Unmarshal([]byte(queryStr), &values); err != nil {
		return nil, err
	}
	if len(values) != 2 {
		return nil, errors.New("two nums expected")
	}

	offset := values[0]
	limit := values[1] - values[0] + 1

	if offset < 0 || limit < 0 {
		return nil, errors.New("invalid offset or limit")
	}
	return &GetAllLinksPagination{Limit: values[1] - values[0] + 1, Offset: values[0]}, nil
}

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
	}

	resultQuery := strings.ToLower(values[0]) + " " + strings.ToLower(values[1])

	val, ok := orderByMap[resultQuery]
	if !ok {
		return "id ASC"
	}
	return val
}
