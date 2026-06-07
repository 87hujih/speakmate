package state

import (
	"context"
	"errors"
	"time"

	"speakmate/internal/model"
)

const (
	// DefaultSessionStateTTL 是训练临时上下文、状态、评分和纠错摘要的默认保留时间。
	DefaultSessionStateTTL = 2 * time.Hour
	// DefaultWebSocketConnectionTTL 是 WebSocket 连接状态的默认保留时间。
	DefaultWebSocketConnectionTTL = 30 * time.Minute
)

var (
	// ErrStateNotFound 表示短期状态不存在或已经过期。
	ErrStateNotFound = errors.New("session state not found")
	// ErrInvalidState 表示写入短期状态时缺少必要字段。
	ErrInvalidState = errors.New("invalid session state")
)

// SessionStateStore 管理训练过程中的短期状态，不替代长期持久化仓库。
type SessionStateStore interface {
	SaveMessageSnapshot(ctx context.Context, sessionID int, messages []model.Message) error
	GetMessageSnapshot(ctx context.Context, sessionID int) ([]model.Message, error)
	SaveSessionState(ctx context.Context, state SessionState) error
	GetSessionState(ctx context.Context, sessionID int) (SessionState, error)
	SavePartialScore(ctx context.Context, score model.ScoreResult) error
	GetPartialScore(ctx context.Context, sessionID int) (model.ScoreResult, error)
	AppendCorrection(ctx context.Context, correction model.CorrectionResult) error
	ListCorrections(ctx context.Context, sessionID int) ([]model.CorrectionResult, error)
	SaveWebSocketConnection(ctx context.Context, connection WebSocketConnectionState) error
	GetWebSocketConnection(ctx context.Context, sessionID int) (WebSocketConnectionState, error)
}

// SessionState 是当前训练阶段、轮次和生命周期状态的短期快照。
type SessionState struct {
	SessionID  int       `json:"session_id"`
	ScenarioID int       `json:"scenario_id"`
	UserID     int       `json:"user_id"`
	Status     string    `json:"status"`
	Stage      string    `json:"stage"`
	TurnCount  int       `json:"turn_count"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WebSocketConnectionState 是实时音频 WebSocket 连接的短期状态。
type WebSocketConnectionState struct {
	SessionID    int       `json:"session_id"`
	Status       string    `json:"status"`
	ContentType  string    `json:"content_type,omitempty"`
	ChunkCount   int       `json:"chunk_count,omitempty"`
	LastSequence int       `json:"last_sequence,omitempty"`
	LastError    string    `json:"last_error,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}
