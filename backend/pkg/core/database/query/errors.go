package query

import (
	"errors"

	"github.com/uptrace/bun/driver/pgdriver"
)

// IsUniqueViolation reports a Postgres unique_violation (23505), e.g. a concurrent insert that
// slipped past an existence check.
func IsUniqueViolation(err error) bool {
	var pgErr pgdriver.Error
	return errors.As(err, &pgErr) && pgErr.Field('C') == "23505"
}

// UniqueViolationConstraint returns the name of the unique index or constraint a Postgres
// unique_violation hit, or "" for any other error.
func UniqueViolationConstraint(err error) string {
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) && pgErr.Field('C') == "23505" {
		return pgErr.Field('n')
	}
	return ""
}
