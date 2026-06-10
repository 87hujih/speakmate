package model

import "time"

// SessionListQuery 表示训练历史列表查询条件。
type SessionListQuery struct {
	UserID   int
	Page     int
	PageSize int
}

// SessionWindowQuery 表示按创建时间窗口查询训练 Session 的条件。
type SessionWindowQuery struct {
	UserID    int
	StartedAt time.Time
	EndedAt   time.Time
	Limit     int
}

// SessionListResult 表示训练历史列表仓库查询结果。
type SessionListResult struct {
	Sessions []Session
	Total    int
}
