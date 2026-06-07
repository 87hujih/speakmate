package stream

import (
	"errors"
	"sync"
	"time"
)

const defaultBufferSize = 16

var (
	// ErrBusClosed 表示事件总线已经关闭，不能继续订阅或发布。
	ErrBusClosed = errors.New("stream bus closed")
)

// Bus 是按 session_id 隔离的内存事件总线。
type Bus struct {
	mu          sync.RWMutex
	bufferSize  int
	closed      bool
	subscribers map[int]map[chan Event]struct{}
}

// NewBus 创建内存事件总线。
func NewBus() *Bus {
	return &Bus{
		bufferSize:  defaultBufferSize,
		subscribers: make(map[int]map[chan Event]struct{}),
	}
}

// Subscribe 订阅某个 Session 的事件。
func (b *Bus) Subscribe(sessionID int) (<-chan Event, func(), error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, nil, ErrBusClosed
	}

	events := make(chan Event, b.bufferSize)
	if b.subscribers[sessionID] == nil {
		b.subscribers[sessionID] = make(map[chan Event]struct{})
	}
	b.subscribers[sessionID][events] = struct{}{}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.unsubscribe(sessionID, events)
		})
	}

	return events, unsubscribe, nil
}

// Publish 向事件所属 Session 的所有订阅者发布事件。
func (b *Bus) Publish(event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrBusClosed
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	for subscriber := range b.subscribers[event.SessionID] {
		select {
		case subscriber <- event:
		default:
		}
	}

	return nil
}

// Close 关闭事件总线和全部订阅。
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true
	for sessionID, subscribers := range b.subscribers {
		for subscriber := range subscribers {
			close(subscriber)
			delete(subscribers, subscriber)
		}
		delete(b.subscribers, sessionID)
	}
}

func (b *Bus) unsubscribe(sessionID int, events chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscribers, ok := b.subscribers[sessionID]
	if !ok {
		return
	}
	if _, ok := subscribers[events]; !ok {
		return
	}

	close(events)
	delete(subscribers, events)
	if len(subscribers) == 0 {
		delete(b.subscribers, sessionID)
	}
}
