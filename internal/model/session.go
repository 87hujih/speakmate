package model

import "time"

// SessionStatus 表示训练 Session 当前生命周期状态。
type SessionStatus string

const (
	// SessionStatusRunning 表示训练正在进行中。
	SessionStatusRunning SessionStatus = "running"
	// SessionStatusFinished 表示训练已经结束。
	SessionStatusFinished SessionStatus = "finished"
)

// Session 表示一次口语训练生命周期。
type Session struct {
	// ID 是 Session 的内部数字标识。
	ID int `json:"session_id"`
	// SessionNo 是展示或排查用的稳定编号。
	SessionNo string `json:"session_no"`
	// ScenarioID 是该 Session 关联的训练场景 ID。
	ScenarioID int `json:"scenario_id"`
	// UserID 是发起训练的用户 ID；当前无登录系统时默认使用 1。
	UserID int `json:"user_id"`
	// Status 是 Session 当前生命周期状态。
	Status SessionStatus `json:"status"`
	// TurnCount 是当前对话轮次，Module 2 阶段固定从 0 开始。
	TurnCount int `json:"turn_count"`
	// Messages 是 Session 下的消息列表，Module 2 阶段通常为空。
	Messages []Message `json:"messages"`
	// CreatedAt 是 Session 创建时间。
	CreatedAt time.Time `json:"created_at"`
	// EndedAt 是 Session 结束时间，未结束时为空。
	EndedAt *time.Time `json:"ended_at"`
}
