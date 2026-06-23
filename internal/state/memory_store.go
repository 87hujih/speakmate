package state

import (
	"context"
	"sync"
	"time"

	"speakmate/internal/model"
)

// memoryEntry 保存内存短期状态及过期时间。
type memoryEntry[T any] struct {
	value     T
	expiresAt time.Time
}

// MemorySessionStateStore 是本地开发和自动测试使用的短期状态存储。
type MemorySessionStateStore struct {
	mu           sync.RWMutex
	now          func() time.Time
	stateTTL     time.Duration
	webSocketTTL time.Duration

	messages    map[int]memoryEntry[[]model.Message]
	states      map[int]memoryEntry[SessionState]
	scores      map[int]memoryEntry[model.ScoreResult]
	corrections map[int]memoryEntry[[]model.CorrectionResult]
	connections map[int]memoryEntry[WebSocketConnectionState]
}

// MemoryStoreOption 用于配置内存短期状态存储。
type MemoryStoreOption func(*MemorySessionStateStore)

// NewMemorySessionStateStore 创建内存短期状态存储。
func NewMemorySessionStateStore(opts ...MemoryStoreOption) *MemorySessionStateStore {
	store := &MemorySessionStateStore{
		now:          time.Now,
		stateTTL:     DefaultSessionStateTTL,
		webSocketTTL: DefaultWebSocketConnectionTTL,
		messages:     make(map[int]memoryEntry[[]model.Message]),
		states:       make(map[int]memoryEntry[SessionState]),
		scores:       make(map[int]memoryEntry[model.ScoreResult]),
		corrections:  make(map[int]memoryEntry[[]model.CorrectionResult]),
		connections:  make(map[int]memoryEntry[WebSocketConnectionState]),
	}
	for _, opt := range opts {
		opt(store)
	}
	if store.now == nil {
		store.now = time.Now
	}
	if store.stateTTL <= 0 {
		store.stateTTL = DefaultSessionStateTTL
	}
	if store.webSocketTTL <= 0 {
		store.webSocketTTL = DefaultWebSocketConnectionTTL
	}

	return store
}

// WithMemoryClock 注入时间函数，便于 TTL 测试。
func WithMemoryClock(now func() time.Time) MemoryStoreOption {
	return func(store *MemorySessionStateStore) {
		if now != nil {
			store.now = now
		}
	}
}

// WithMemoryStateTTL 覆盖训练临时状态 TTL。
func WithMemoryStateTTL(ttl time.Duration) MemoryStoreOption {
	return func(store *MemorySessionStateStore) {
		if ttl > 0 {
			store.stateTTL = ttl
		}
	}
}

// WithMemoryWebSocketTTL 覆盖 WebSocket 连接状态 TTL。
func WithMemoryWebSocketTTL(ttl time.Duration) MemoryStoreOption {
	return func(store *MemorySessionStateStore) {
		if ttl > 0 {
			store.webSocketTTL = ttl
		}
	}
}

// SaveMessageSnapshot 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) SaveMessageSnapshot(ctx context.Context, sessionID int, messages []model.Message) error {
	if sessionID <= 0 {
		return ErrInvalidState
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[sessionID] = memoryEntry[[]model.Message]{
		value:     cloneMessages(messages),
		expiresAt: s.expiresAt(s.stateTTL),
	}

	return nil
}

// GetMessageSnapshot 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) GetMessageSnapshot(ctx context.Context, sessionID int) ([]model.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.messages[sessionID]
	if !ok || s.entryExpired(entry.expiresAt) {
		delete(s.messages, sessionID)
		return nil, ErrStateNotFound
	}

	return cloneMessages(entry.value), nil
}

// SaveSessionState 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) SaveSessionState(ctx context.Context, state SessionState) error {
	if state.SessionID <= 0 {
		return ErrInvalidState
	}
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = s.now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state.SessionID] = memoryEntry[SessionState]{
		value:     state,
		expiresAt: s.expiresAt(s.stateTTL),
	}

	return nil
}

// GetSessionState 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) GetSessionState(ctx context.Context, sessionID int) (SessionState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.states[sessionID]
	if !ok || s.entryExpired(entry.expiresAt) {
		delete(s.states, sessionID)
		return SessionState{}, ErrStateNotFound
	}

	return entry.value, nil
}

// SavePartialScore 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) SavePartialScore(ctx context.Context, score model.ScoreResult) error {
	if score.SessionID <= 0 {
		return ErrInvalidState
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[score.SessionID] = memoryEntry[model.ScoreResult]{
		value:     score,
		expiresAt: s.expiresAt(s.stateTTL),
	}

	return nil
}

// GetPartialScore 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) GetPartialScore(ctx context.Context, sessionID int) (model.ScoreResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.scores[sessionID]
	if !ok || s.entryExpired(entry.expiresAt) {
		delete(s.scores, sessionID)
		return model.ScoreResult{}, ErrStateNotFound
	}

	return entry.value, nil
}

// AppendCorrection 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) AppendCorrection(ctx context.Context, correction model.CorrectionResult) error {
	if correction.SessionID <= 0 {
		return ErrInvalidState
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	existing := []model.CorrectionResult{}
	if entry, ok := s.corrections[correction.SessionID]; ok && !s.entryExpired(entry.expiresAt) {
		existing = cloneCorrections(entry.value)
	}
	existing = append(existing, cloneCorrection(correction))
	s.corrections[correction.SessionID] = memoryEntry[[]model.CorrectionResult]{
		value:     existing,
		expiresAt: s.expiresAt(s.stateTTL),
	}

	return nil
}

// ListCorrections 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) ListCorrections(ctx context.Context, sessionID int) ([]model.CorrectionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.corrections[sessionID]
	if !ok || s.entryExpired(entry.expiresAt) {
		delete(s.corrections, sessionID)
		return nil, ErrStateNotFound
	}

	return cloneCorrections(entry.value), nil
}

// SaveWebSocketConnection 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) SaveWebSocketConnection(ctx context.Context, connection WebSocketConnectionState) error {
	if connection.SessionID <= 0 {
		return ErrInvalidState
	}
	if connection.UpdatedAt.IsZero() {
		connection.UpdatedAt = s.now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[connection.SessionID] = memoryEntry[WebSocketConnectionState]{
		value:     connection,
		expiresAt: s.expiresAt(s.webSocketTTL),
	}

	return nil
}

// GetWebSocketConnection 封装当前文件中的辅助处理逻辑。
func (s *MemorySessionStateStore) GetWebSocketConnection(ctx context.Context, sessionID int) (WebSocketConnectionState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.connections[sessionID]
	if !ok || s.entryExpired(entry.expiresAt) {
		delete(s.connections, sessionID)
		return WebSocketConnectionState{}, ErrStateNotFound
	}

	return entry.value, nil
}

// expiresAt 根据 TTL 计算过期时间。
func (s *MemorySessionStateStore) expiresAt(ttl time.Duration) time.Time {
	return s.now().UTC().Add(ttl)
}

// entryExpired 判断内存状态条目是否已过期。
func (s *MemorySessionStateStore) entryExpired(expiresAt time.Time) bool {
	return !expiresAt.IsZero() && !s.now().UTC().Before(expiresAt)
}

// cloneMessages 复制消息列表，避免外部修改状态缓存。
func cloneMessages(messages []model.Message) []model.Message {
	if messages == nil {
		return nil
	}
	cloned := make([]model.Message, len(messages))
	copy(cloned, messages)

	return cloned
}

// cloneCorrections 复制纠错列表，避免外部修改状态缓存。
func cloneCorrections(corrections []model.CorrectionResult) []model.CorrectionResult {
	if corrections == nil {
		return nil
	}
	cloned := make([]model.CorrectionResult, len(corrections))
	for i, correction := range corrections {
		cloned[i] = cloneCorrection(correction)
	}

	return cloned
}

// cloneCorrection 复制单条纠错结果。
func cloneCorrection(correction model.CorrectionResult) model.CorrectionResult {
	if correction.Errors != nil {
		correction.Errors = append([]model.CorrectionError(nil), correction.Errors...)
	}
	if correction.BetterExpressions != nil {
		correction.BetterExpressions = append([]string(nil), correction.BetterExpressions...)
	}

	return correction
}
