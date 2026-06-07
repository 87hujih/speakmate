package model

// SessionListQuery 表示训练历史列表查询条件。
type SessionListQuery struct {
	UserID   int
	Page     int
	PageSize int
}

// SessionListResult 表示训练历史列表仓库查询结果。
type SessionListResult struct {
	Sessions []Session
	Total    int
}
