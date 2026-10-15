package links

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestLinkWriteError(t *testing.T) {
	if !errors.Is(linkWriteError(fmt.Errorf("update: %w", sql.ErrNoRows)), ErrLinkNotFound) {
		t.Error("missing update row was not mapped to ErrLinkNotFound")
	}
	for _, constraint := range []string{"links_short_name_key", "links_short_url_key"} {
		err := fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505", ConstraintName: constraint})
		if !errors.Is(linkWriteError(err), ErrShortNameInUse) {
			t.Errorf("%s was not mapped to a short name conflict", constraint)
		}
	}
	for _, err := range []error{
		errors.New("connection failure"),
		&pgconn.PgError{Code: "23503", ConstraintName: "links_short_name_key"},
		&pgconn.PgError{Code: "23505", ConstraintName: "links_pkey"},
	} {
		if linkWriteError(err) != err {
			t.Errorf("unrelated error was changed: %v", err)
		}
	}
}
