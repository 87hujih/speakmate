package handler

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/stream"
)

func TestStreamHandlerWritesSessionEvents(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	subscriber := newFakeEventSubscriber()
	streamHandler := NewStreamHandler(subscriber, WithStreamHeartbeatInterval(time.Hour))
	engine := gin.New()
	engine.GET("/api/v1/sessions/:id/stream", streamHandler.Stream)
	server := httptest.NewServer(engine)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/sessions/7/stream")
	if err != nil {
		t.Fatalf("GET stream returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("content-type = %q, want text/event-stream", got)
	}
	if subscriber.sessionID != 7 {
		t.Fatalf("subscribed session id = %d, want 7", subscriber.sessionID)
	}

	subscriber.events <- stream.Event{
		Type:      stream.EventTypeAIMessageDone,
		SessionID: 7,
		Payload: map[string]string{
			"content": "hello",
		},
		CreatedAt: time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC),
	}

	reader := bufio.NewReader(resp.Body)
	eventLine := readSSELine(t, reader)
	if eventLine != "event: ai_message_done" {
		t.Fatalf("event line = %q, want ai_message_done", eventLine)
	}
	dataLine := readSSELine(t, reader)
	if !strings.HasPrefix(dataLine, "data: ") {
		t.Fatalf("data line = %q, want data prefix", dataLine)
	}

	var event stream.Event
	if err := json.Unmarshal([]byte(strings.TrimPrefix(dataLine, "data: ")), &event); err != nil {
		t.Fatalf("event data is not valid JSON: %v", err)
	}
	if event.Type != stream.EventTypeAIMessageDone {
		t.Fatalf("event type = %q, want %q", event.Type, stream.EventTypeAIMessageDone)
	}
	if event.SessionID != 7 {
		t.Fatalf("event session id = %d, want 7", event.SessionID)
	}

	_ = resp.Body.Close()
	select {
	case <-subscriber.unsubscribed:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("subscription was not closed after client disconnect")
	}
}

func TestStreamHandlerWritesHeartbeat(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	subscriber := newFakeEventSubscriber()
	streamHandler := NewStreamHandler(subscriber, WithStreamHeartbeatInterval(10*time.Millisecond))
	engine := gin.New()
	engine.GET("/api/v1/sessions/:id/stream", streamHandler.Stream)
	server := httptest.NewServer(engine)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/sessions/7/stream")
	if err != nil {
		t.Fatalf("GET stream returned error: %v", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	line := readSSELine(t, reader)
	if line != ": ping" {
		t.Fatalf("heartbeat line = %q, want : ping", line)
	}
}

func TestStreamHandlerRejectsInvalidSessionID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	subscriber := newFakeEventSubscriber()
	streamHandler := NewStreamHandler(subscriber)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:id/stream", streamHandler.Stream)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/abc/stream", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertSessionErrorResponse(t, rec, http.StatusBadRequest, invalidSessionIDCode, "训练 ID 无效")
	if subscriber.callCount != 0 {
		t.Fatalf("subscribe call count = %d, want 0", subscriber.callCount)
	}
}

func readSSELine(t *testing.T, reader *bufio.Reader) string {
	t.Helper()

	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("ReadString returned error: %v", err)
	}

	return strings.TrimRight(line, "\r\n")
}

type fakeEventSubscriber struct {
	events       chan stream.Event
	unsubscribed chan struct{}
	sessionID    int
	callCount    int
}

func newFakeEventSubscriber() *fakeEventSubscriber {
	return &fakeEventSubscriber{
		events:       make(chan stream.Event, 4),
		unsubscribed: make(chan struct{}),
	}
}

func (s *fakeEventSubscriber) Subscribe(sessionID int) (<-chan stream.Event, func(), error) {
	s.callCount++
	s.sessionID = sessionID

	return s.events, func() {
		close(s.unsubscribed)
	}, nil
}

func assertSessionErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, httpStatus int, code int, message string) {
	t.Helper()

	if rec.Code != httpStatus {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, httpStatus, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Code != code {
		t.Fatalf("code = %d, want %d", body.Code, code)
	}
	if body.Message != message {
		t.Fatalf("message = %q, want %q", body.Message, message)
	}
}
