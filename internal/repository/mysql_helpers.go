package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// scanner 抽象 sql.Row 和 sql.Rows 的 Scan 能力。
type scanner interface {
	Scan(dest ...any) error
}

// marshalJSON 将结构化字段序列化为数据库 JSON 字符串。
func marshalJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// unmarshalJSON 将数据库 JSON 字符串反序列化为结构化字段。
func unmarshalJSON(raw string, target any) error {
	if raw == "" {
		raw = "null"
	}

	return json.Unmarshal([]byte(raw), target)
}

// nullableTimePtr 将可空时间指针转换为 sql.NullTime。
func nullableTimePtr(value *time.Time) any {
	if value == nil {
		return nil
	}

	return *value
}

// timePtrFromNull 将 sql.NullTime 转换为可空时间指针。
func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	t := value.Time
	return &t
}

// notFoundFromNoRows 将 sql.ErrNoRows 映射为指定仓库错误。
func notFoundFromNoRows(err error, notFound error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}

	return err
}
