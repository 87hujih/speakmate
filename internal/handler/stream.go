package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/response"
	"speakmate/internal/stream"
)

const (
	streamUnavailableCode = 7001
	defaultHeartbeat      = 15 * time.Second
)

// EventSubscriber 定义 SSE Handler 依赖的事件订阅能力。
type EventSubscriber interface {
	Subscribe(sessionID int) (<-chan stream.Event, func(), error)
}

// StreamHandler 负责处理 Session SSE 连接。
type StreamHandler struct {
	subscriber        EventSubscriber
	heartbeatInterval time.Duration
}

type StreamOption func(*StreamHandler)

// NewStreamHandler 创建 SSE Handler。
func NewStreamHandler(subscriber EventSubscriber, opts ...StreamOption) *StreamHandler {
	handler := &StreamHandler{
		subscriber:        subscriber,
		heartbeatInterval: defaultHeartbeat,
	}
	for _, opt := range opts {
		opt(handler)
	}
	if handler.heartbeatInterval <= 0 {
		handler.heartbeatInterval = defaultHeartbeat
	}

	return handler
}

// WithStreamHeartbeatInterval 设置 SSE 心跳间隔。
func WithStreamHeartbeatInterval(interval time.Duration) StreamOption {
	return func(handler *StreamHandler) {
		handler.heartbeatInterval = interval
	}
}

// Stream 建立 Session 级 SSE 连接。
func (h *StreamHandler) Stream(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}
	if h.subscriber == nil {
		response.Error(c, http.StatusServiceUnavailable, streamUnavailableCode, "stream unavailable")
		return
	}

	events, unsubscribe, err := h.subscriber.Subscribe(id)
	if err != nil {
		writeStreamError(c, err)
		return
	}
	defer unsubscribe()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	header := c.Writer.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher.Flush()

	ticker := time.NewTicker(h.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeSSEEvent(c.Writer, flusher, event); err != nil {
				return
			}
		case <-ticker.C:
			if err := writeSSEHeartbeat(c.Writer, flusher); err != nil {
				return
			}
		}
	}
}

func writeStreamError(c *gin.Context, err error) {
	if errors.Is(err, stream.ErrBusClosed) {
		response.Error(c, http.StatusServiceUnavailable, streamUnavailableCode, "stream unavailable")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
}

func writeSSEEvent(writer io.Writer, flusher http.Flusher, event stream.Event) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(writer, "event: "+string(event.Type)+"\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, "data: "+string(data)+"\n\n"); err != nil {
		return err
	}
	flusher.Flush()

	return nil
}

func writeSSEHeartbeat(writer io.Writer, flusher http.Flusher) error {
	if _, err := io.WriteString(writer, ": ping\n\n"); err != nil {
		return err
	}
	flusher.Flush()

	return nil
}
