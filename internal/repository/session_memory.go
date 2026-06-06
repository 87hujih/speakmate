package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"speakmate/internal/model"
)

var (
	// ErrSessionNotFound 表示内存仓库中没有找到对应 Session。
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionAlreadyFinished 表示 Session 已经结束，不能再次结束。
	ErrSessionAlreadyFinished = errors.New("session already finished")
)

// MemorySessionRepository 使用内存 map 保存训练 Session。
type MemorySessionRepository struct {
	mu       sync.RWMutex
	nextID   int
	sessions map[int]model.Session
}

// NewMemorySessionRepository 创建空的内存 Session 仓库。
func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		nextID:   1,
		sessions: make(map[int]model.Session),
	}
}

// Create 保存新的 Session，并生成 session_id 和 session_no。
func (r *MemorySessionRepository) Create(session model.Session) (model.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session.ID = r.nextID
	r.nextID++
	if session.SessionNo == "" {
		session.SessionNo = fmt.Sprintf("S%s%04d", session.CreatedAt.Format("20060102"), session.ID)
	}
	if session.Messages == nil {
		session.Messages = []model.Message{}
	}

	r.sessions[session.ID] = cloneSession(session)

	return cloneSession(session), nil
}

// FindByID 按内部数字 ID 查询 Session。
func (r *MemorySessionRepository) FindByID(id int) (model.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return model.Session{}, ErrSessionNotFound
	}

	return cloneSession(session), nil
}

// Finish 把 running Session 原子地结束为 finished。
func (r *MemorySessionRepository) Finish(id int, endedAt time.Time) (model.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return model.Session{}, ErrSessionNotFound
	}
	if session.Status == model.SessionStatusFinished {
		return model.Session{}, ErrSessionAlreadyFinished
	}

	session.Status = model.SessionStatusFinished
	session.EndedAt = &endedAt
	r.sessions[id] = cloneSession(session)

	return cloneSession(session), nil
}

func cloneSession(session model.Session) model.Session {
	if session.Messages != nil {
		session.Messages = append([]model.Message(nil), session.Messages...)
	}
	if session.EndedAt != nil {
		endedAt := *session.EndedAt
		session.EndedAt = &endedAt
	}

	return session
}
