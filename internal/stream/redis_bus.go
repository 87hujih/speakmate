package stream

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// 事件流模块使用的事件类型和默认值。
const (
	defaultRedisBusBufferSize = 16
	defaultRedisBusEventLimit = 200
	// DefaultRedisBusEventTTL 是 Redis 事件留存列表的默认保留时间。
	DefaultRedisBusEventTTL = 30 * time.Minute
)

// RedisBus 使用 Redis Pub/Sub 分发 SSE 事件，并用 List 保留短期事件副本。
type RedisBus struct {
	client     *goredis.Client
	eventTTL   time.Duration
	bufferSize int
	eventLimit int64
}

// RedisBusOption 用于配置 Redis 事件总线。
type RedisBusOption func(*RedisBus)

// NewRedisBus 创建 Redis 事件总线。
func NewRedisBus(client *goredis.Client, opts ...RedisBusOption) *RedisBus {
	bus := &RedisBus{
		client:     client,
		eventTTL:   DefaultRedisBusEventTTL,
		bufferSize: defaultRedisBusBufferSize,
		eventLimit: defaultRedisBusEventLimit,
	}
	for _, opt := range opts {
		opt(bus)
	}
	if bus.eventTTL <= 0 {
		bus.eventTTL = DefaultRedisBusEventTTL
	}
	if bus.bufferSize <= 0 {
		bus.bufferSize = defaultRedisBusBufferSize
	}
	if bus.eventLimit <= 0 {
		bus.eventLimit = defaultRedisBusEventLimit
	}

	return bus
}

// WithRedisBusEventTTL 覆盖事件留存 TTL。
func WithRedisBusEventTTL(ttl time.Duration) RedisBusOption {
	return func(bus *RedisBus) {
		if ttl > 0 {
			bus.eventTTL = ttl
		}
	}
}

// Subscribe 订阅指定 Session 的 Redis Pub/Sub 事件。
func (b *RedisBus) Subscribe(sessionID int) (<-chan Event, func(), error) {
	if b == nil || b.client == nil {
		return nil, nil, ErrBusClosed
	}
	channel := sessionEventsKey(sessionID)
	pubsub := b.client.Subscribe(context.Background(), channel)
	if _, err := pubsub.Receive(context.Background()); err != nil {
		_ = pubsub.Close()
		return nil, nil, err
	}

	events := make(chan Event, b.bufferSize)
	done := make(chan struct{})
	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			close(done)
			_ = pubsub.Close()
		})
	}

	go func() {
		defer close(events)
		messages := pubsub.Channel()
		for {
			select {
			case <-done:
				return
			case message, ok := <-messages:
				if !ok {
					return
				}
				var event Event
				if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
					continue
				}
				select {
				case events <- event:
				default:
				}
			}
		}
	}()

	return events, unsubscribe, nil
}

// Publish 发布事件到 Redis Pub/Sub，并写入带 TTL 的短期事件列表。
func (b *RedisBus) Publish(event Event) error {
	if b == nil || b.client == nil {
		return ErrBusClosed
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	key := sessionEventsKey(event.SessionID)
	pipe := b.client.TxPipeline()
	pipe.RPush(context.Background(), key, payload)
	pipe.LTrim(context.Background(), key, -b.eventLimit, -1)
	pipe.Expire(context.Background(), key, b.eventTTL)
	pipe.Publish(context.Background(), key, payload)
	_, err = pipe.Exec(context.Background())

	return err
}

// sessionEventsKey 构造 Redis 事件列表键。
func sessionEventsKey(sessionID int) string {
	return "session:" + strconv.Itoa(sessionID) + ":events"
}
