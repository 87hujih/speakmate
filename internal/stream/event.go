package stream

import "time"

// EventType 表示 SSE 推送给前端的业务事件类型。
type EventType string

// 事件流模块使用的事件类型和默认值。
const (
	// EventTypeAIMessageDelta 表示 AI 回复的一个文本片段。
	EventTypeAIMessageDelta EventType = "ai_message_delta"
	// EventTypeAIMessageDone 表示 AI 完整回复已经生成并保存。
	EventTypeAIMessageDone EventType = "ai_message_done"
	// EventTypeCorrectionDone 表示单条用户消息纠错已经完成。
	EventTypeCorrectionDone EventType = "correction_done"
	// EventTypeScoreUpdated 表示 Session 当前评分已经更新。
	EventTypeScoreUpdated EventType = "score_updated"
	// EventTypeReportDone 表示课后报告已经生成。
	EventTypeReportDone EventType = "report_done"
	// EventTypeError 表示某个异步可见步骤失败。
	EventTypeError EventType = "error"
)

// Event 是 session 级 SSE 事件的统一结构。
type Event struct {
	Type      EventType `json:"type"`
	SessionID int       `json:"session_id"`
	Payload   any       `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// AIMessageDeltaPayload 是 AI 回复分片事件的数据。
type AIMessageDeltaPayload struct {
	MessageID int    `json:"message_id"`
	Delta     string `json:"delta"`
}

// AIMessageDonePayload 是 AI 完整回复完成事件的数据。
type AIMessageDonePayload struct {
	MessageID int    `json:"message_id"`
	Content   string `json:"content"`
	Stage     string `json:"stage"`
}

// CorrectionDonePayload 是纠错完成事件的数据。
type CorrectionDonePayload struct {
	MessageID  int  `json:"message_id"`
	HasErrors  bool `json:"has_errors"`
	ErrorCount int  `json:"error_count"`
}

// ScoreUpdatedPayload 是评分更新事件的数据。
type ScoreUpdatedPayload struct {
	MessageID  int `json:"message_id"`
	TotalScore int `json:"total_score"`
	Grammar    int `json:"grammar"`
	Expression int `json:"expression"`
}

// ReportDonePayload 是报告生成完成事件的数据。
type ReportDonePayload struct {
	TotalScore int    `json:"total_score"`
	Summary    string `json:"summary"`
}

// ErrorPayload 是错误事件的数据。
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
