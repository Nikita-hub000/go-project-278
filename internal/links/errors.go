package links

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrShortNameInUse = errors.New("short name already in use")

// Only constraints derived from the short name represent a user input conflict.
func linkWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "links_short_name_key", "links_short_url_key":
			return ErrShortNameInUse
		}
	}
	return err
}
