package state

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"speakmate/internal/model"
)

// RedisSessionStateStore 使用 Redis 管理训练短期状态。
type RedisSessionStateStore struct {
	client       *goredis.Client
	stateTTL     time.Duration
	webSocketTTL time.Duration
}

// RedisStoreOption 用于配置 Redis 短期状态存储。
type RedisStoreOption func(*RedisSessionStateStore)

// NewRedisSessionStateStore 创建 Redis 短期状态存储。
func NewRedisSessionStateStore(client *goredis.Client, opts ...RedisStoreOption) *RedisSessionStateStore {
	store := &RedisSessionStateStore{
		client:       client,
		stateTTL:     DefaultSessionStateTTL,
		webSocketTTL: DefaultWebSocketConnectionTTL,
	}
	for _, opt := range opts {
		opt(store)
	}
	if store.stateTTL <= 0 {
		store.stateTTL = DefaultSessionStateTTL
	}
	if store.webSocketTTL <= 0 {
		store.webSocketTTL = DefaultWebSocketConnectionTTL
	}

	return store
}

// WithRedisStateTTL 覆盖训练临时状态 TTL。
func WithRedisStateTTL(ttl time.Duration) RedisStoreOption {
	return func(store *RedisSessionStateStore) {
		if ttl > 0 {
			store.stateTTL = ttl
		}
	}
}

// WithRedisWebSocketTTL 覆盖 WebSocket 连接状态 TTL。
func WithRedisWebSocketTTL(ttl time.Duration) RedisStoreOption {
	return func(store *RedisSessionStateStore) {
		if ttl > 0 {
			store.webSocketTTL = ttl
		}
	}
}

// SaveMessageSnapshot 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) SaveMessageSnapshot(ctx context.Context, sessionID int, messages []model.Message) error {
	if sessionID <= 0 || s.client == nil {
		return ErrInvalidState
	}
	key := sessionMessagesKey(sessionID)
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, key)
	for _, message := range messages {
		payload, err := json.Marshal(message)
		if err != nil {
			return err
		}
		pipe.RPush(ctx, key, payload)
	}
	if len(messages) > 0 {
		pipe.Expire(ctx, key, s.stateTTL)
	}
	_, err := pipe.Exec(ctx)

	return err
}

// GetMessageSnapshot 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) GetMessageSnapshot(ctx context.Context, sessionID int) ([]model.Message, error) {
	if sessionID <= 0 || s.client == nil {
		return nil, ErrInvalidState
	}
	values, err := s.client.LRange(ctx, sessionMessagesKey(sessionID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrStateNotFound
	}
	messages := make([]model.Message, 0, len(values))
	for _, value := range values {
		var message model.Message
		if err := json.Unmarshal([]byte(value), &message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// SaveSessionState 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) SaveSessionState(ctx context.Context, state SessionState) error {
	if state.SessionID <= 0 || s.client == nil {
		return ErrInvalidState
	}
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = time.Now().UTC()
	}
	key := sessionStateKey(state.SessionID)
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, map[string]any{
		"session_id":  state.SessionID,
		"scenario_id": state.ScenarioID,
		"user_id":     state.UserID,
		"status":      state.Status,
		"stage":       state.Stage,
		"turn_count":  state.TurnCount,
		"updated_at":  state.UpdatedAt.UTC().Format(time.RFC3339Nano),
	})
	pipe.Expire(ctx, key, s.stateTTL)
	_, err := pipe.Exec(ctx)

	return err
}

// GetSessionState 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) GetSessionState(ctx context.Context, sessionID int) (SessionState, error) {
	if sessionID <= 0 || s.client == nil {
		return SessionState{}, ErrInvalidState
	}
	values, err := s.client.HGetAll(ctx, sessionStateKey(sessionID)).Result()
	if err != nil {
		return SessionState{}, err
	}
	if len(values) == 0 {
		return SessionState{}, ErrStateNotFound
	}
	updatedAt, _ := time.Parse(time.RFC3339Nano, values["updated_at"])

	return SessionState{
		SessionID:  intField(values, "session_id"),
		ScenarioID: intField(values, "scenario_id"),
		UserID:     intField(values, "user_id"),
		Status:     values["status"],
		Stage:      values["stage"],
		TurnCount:  intField(values, "turn_count"),
		UpdatedAt:  updatedAt,
	}, nil
}

// SavePartialScore 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) SavePartialScore(ctx context.Context, score model.ScoreResult) error {
	if score.SessionID <= 0 || s.client == nil {
		return ErrInvalidState
	}
	key := sessionPartialScoreKey(score.SessionID)
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, map[string]any{
		"message_id":  score.MessageID,
		"session_id":  score.SessionID,
		"fluency":     score.Fluency,
		"grammar":     score.Grammar,
		"expression":  score.Expression,
		"vocabulary":  score.Vocabulary,
		"completion":  score.Completion,
		"total_score": score.TotalScore,
		"comment":     score.Comment,
	})
	pipe.Expire(ctx, key, s.stateTTL)
	_, err := pipe.Exec(ctx)

	return err
}

// GetPartialScore 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) GetPartialScore(ctx context.Context, sessionID int) (model.ScoreResult, error) {
	if sessionID <= 0 || s.client == nil {
		return model.ScoreResult{}, ErrInvalidState
	}
	values, err := s.client.HGetAll(ctx, sessionPartialScoreKey(sessionID)).Result()
	if err != nil {
		return model.ScoreResult{}, err
	}
	if len(values) == 0 {
		return model.ScoreResult{}, ErrStateNotFound
	}

	return model.ScoreResult{
		MessageID:  intField(values, "message_id"),
		SessionID:  intField(values, "session_id"),
		Fluency:    intField(values, "fluency"),
		Grammar:    intField(values, "grammar"),
		Expression: intField(values, "expression"),
		Vocabulary: intField(values, "vocabulary"),
		Completion: intField(values, "completion"),
		TotalScore: intField(values, "total_score"),
		Comment:    values["comment"],
	}, nil
}

// AppendCorrection 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) AppendCorrection(ctx context.Context, correction model.CorrectionResult) error {
	if correction.SessionID <= 0 || s.client == nil {
		return ErrInvalidState
	}
	payload, err := json.Marshal(correction)
	if err != nil {
		return err
	}
	key := sessionCorrectionsKey(correction.SessionID)
	pipe := s.client.TxPipeline()
	pipe.RPush(ctx, key, payload)
	pipe.Expire(ctx, key, s.stateTTL)
	_, err = pipe.Exec(ctx)

	return err
}

// ListCorrections 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) ListCorrections(ctx context.Context, sessionID int) ([]model.CorrectionResult, error) {
	if sessionID <= 0 || s.client == nil {
		return nil, ErrInvalidState
	}
	values, err := s.client.LRange(ctx, sessionCorrectionsKey(sessionID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrStateNotFound
	}
	corrections := make([]model.CorrectionResult, 0, len(values))
	for _, value := range values {
		var correction model.CorrectionResult
		if err := json.Unmarshal([]byte(value), &correction); err != nil {
			return nil, err
		}
		corrections = append(corrections, correction)
	}

	return corrections, nil
}

// SaveWebSocketConnection 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) SaveWebSocketConnection(ctx context.Context, connection WebSocketConnectionState) error {
	if connection.SessionID <= 0 || s.client == nil {
		return ErrInvalidState
	}
	if connection.UpdatedAt.IsZero() {
		connection.UpdatedAt = time.Now().UTC()
	}
	key := webSocketConnectionKey(connection.SessionID)
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, key, map[string]any{
		"session_id":    connection.SessionID,
		"status":        connection.Status,
		"content_type":  connection.ContentType,
		"chunk_count":   connection.ChunkCount,
		"last_sequence": connection.LastSequence,
		"last_error":    connection.LastError,
		"updated_at":    connection.UpdatedAt.UTC().Format(time.RFC3339Nano),
	})
	pipe.Expire(ctx, key, s.webSocketTTL)
	_, err := pipe.Exec(ctx)

	return err
}

// GetWebSocketConnection 封装当前文件中的辅助处理逻辑。
func (s *RedisSessionStateStore) GetWebSocketConnection(ctx context.Context, sessionID int) (WebSocketConnectionState, error) {
	if sessionID <= 0 || s.client == nil {
		return WebSocketConnectionState{}, ErrInvalidState
	}
	values, err := s.client.HGetAll(ctx, webSocketConnectionKey(sessionID)).Result()
	if err != nil {
		return WebSocketConnectionState{}, err
	}
	if len(values) == 0 {
		return WebSocketConnectionState{}, ErrStateNotFound
	}
	updatedAt, _ := time.Parse(time.RFC3339Nano, values["updated_at"])

	return WebSocketConnectionState{
		SessionID:    intField(values, "session_id"),
		Status:       values["status"],
		ContentType:  values["content_type"],
		ChunkCount:   intField(values, "chunk_count"),
		LastSequence: intField(values, "last_sequence"),
		LastError:    values["last_error"],
		UpdatedAt:    updatedAt,
	}, nil
}

// sessionMessagesKey 构造 Redis 消息快照键。
func sessionMessagesKey(sessionID int) string {
	return "session:" + strconv.Itoa(sessionID) + ":messages"
}

// sessionStateKey 构造 Redis Session 状态键。
func sessionStateKey(sessionID int) string {
	return "session:" + strconv.Itoa(sessionID) + ":state"
}

// sessionPartialScoreKey 构造 Redis 当前评分键。
func sessionPartialScoreKey(sessionID int) string {
	return "session:" + strconv.Itoa(sessionID) + ":partial_score"
}

// sessionCorrectionsKey 构造 Redis 纠错列表键。
func sessionCorrectionsKey(sessionID int) string {
	return "session:" + strconv.Itoa(sessionID) + ":corrections"
}

// webSocketConnectionKey 构造 Redis WebSocket 连接状态键。
func webSocketConnectionKey(sessionID int) string {
	return "ws:" + strconv.Itoa(sessionID) + ":connection"
}

// intField 将 Redis 字段解析为整数。
func intField(values map[string]string, key string) int {
	parsed, _ := strconv.Atoi(values[key])
	return parsed
}
