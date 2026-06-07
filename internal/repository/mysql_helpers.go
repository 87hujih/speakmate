package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type scanner interface {
	Scan(dest ...any) error
}

func marshalJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func unmarshalJSON(raw string, target any) error {
	if raw == "" {
		raw = "null"
	}

	return json.Unmarshal([]byte(raw), target)
}

func nullableTimePtr(value *time.Time) any {
	if value == nil {
		return nil
	}

	return *value
}

func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	t := value.Time
	return &t
}

func notFoundFromNoRows(err error, notFound error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}

	return err
}
