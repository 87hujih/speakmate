package model

import "time"

// MessageRole 表示一条对话消息的发送方角色。
type MessageRole string

const (
	// MessageRoleUser 表示用户发送的消息。
	MessageRoleUser MessageRole = "user"
	// MessageRoleAI 表示 AI 发送的消息。
	MessageRoleAI MessageRole = "ai"
)

// Message 表示训练 Session 中的一条对话消息。
type Message struct {
	// ID 是消息内部数字标识。
	ID int `json:"id"`
	// SessionID 是消息所属的 Session ID。
	SessionID int `json:"session_id"`
	// Role 是消息发送方角色。
	Role MessageRole `json:"role"`
	// Content 是消息文本内容。
	Content string `json:"content"`
	// Stage 是消息所属的训练阶段。
	Stage string `json:"stage"`
	// CreatedAt 是消息创建时间。
	CreatedAt time.Time `json:"created_at"`
}
