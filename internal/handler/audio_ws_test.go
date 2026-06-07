package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"speakmate/internal/config"
	"speakmate/internal/model"
	"speakmate/internal/service"
	"speakmate/internal/state"
)

func TestAudioWebSocketHandlerRecordsErrorAndCloseConnectionState(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	stateStore := newRecordingStateStore()
	audioStreamService := service.NewAudioStreamService(
		nil,
		nil,
		service.WithAudioStreamStateStore(stateStore),
	)
	audioWSHandler := NewAudioWebSocketHandler(audioStreamService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:id/audio/ws", audioWSHandler.Stream)
	server := httptest.NewServer(engine)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsTestURL(server.URL, "/api/v1/sessions/7/audio/ws"), nil)
	if err != nil {
		t.Fatalf("Dial returned error: %v", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte("{")); err != nil {
		t.Fatalf("WriteMessage returned error: %v", err)
	}
	var event audioWSServerEvent
	if err := conn.ReadJSON(&event); err != nil {
		t.Fatalf("ReadJSON returned error: %v", err)
	}
	if event.Type != audioWSEventError {
		t.Fatalf("event type = %q, want %q", event.Type, audioWSEventError)
	}
	waitForConnectionStatus(t, stateStore, 7, "error")

	if err := conn.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	waitForConnectionStatus(t, stateStore, 7, "closed")
}

func TestAudioWebSocketHandlerChecksConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	audioStreamService := service.NewAudioStreamService(nil, nil)
	audioWSHandler := NewAudioWebSocketHandler(audioStreamService, config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
	})
	engine := gin.New()
	engine.GET("/api/v1/sessions/:id/audio/ws", audioWSHandler.Stream)
	server := httptest.NewServer(engine)
	defer server.Close()

	allowedHeader := http.Header{}
	allowedHeader.Set("Origin", "http://localhost:5173")
	allowedConn, _, err := websocket.DefaultDialer.Dial(wsTestURL(server.URL, "/api/v1/sessions/7/audio/ws"), allowedHeader)
	if err != nil {
		t.Fatalf("allowed origin Dial returned error: %v", err)
	}
	if err := allowedConn.Close(); err != nil {
		t.Fatalf("allowed conn Close returned error: %v", err)
	}

	blockedHeader := http.Header{}
	blockedHeader.Set("Origin", "https://evil.example.com")
	blockedConn, resp, err := websocket.DefaultDialer.Dial(wsTestURL(server.URL, "/api/v1/sessions/7/audio/ws"), blockedHeader)
	if err == nil {
		_ = blockedConn.Close()
		t.Fatal("blocked origin Dial succeeded, want bad handshake")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("blocked origin status = %d, want %d", status, http.StatusForbidden)
	}
}

func wsTestURL(serverURL string, path string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http") + path
}

func waitForConnectionStatus(t *testing.T, store *recordingStateStore, sessionID int, status string) {
	t.Helper()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		connection, ok := store.connection(sessionID)
		if ok && connection.Status == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	connection, _ := store.connection(sessionID)
	t.Fatalf("connection status = %+v, want %q", connection, status)
}

type recordingStateStore struct {
	mu          sync.Mutex
	connections map[int]state.WebSocketConnectionState
}

func newRecordingStateStore() *recordingStateStore {
	return &recordingStateStore{
		connections: make(map[int]state.WebSocketConnectionState),
	}
}

func (s *recordingStateStore) connection(sessionID int) (state.WebSocketConnectionState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	connection, ok := s.connections[sessionID]
	return connection, ok
}

func (s *recordingStateStore) SaveMessageSnapshot(ctx context.Context, sessionID int, messages []model.Message) error {
	return nil
}

func (s *recordingStateStore) GetMessageSnapshot(ctx context.Context, sessionID int) ([]model.Message, error) {
	return nil, state.ErrStateNotFound
}

func (s *recordingStateStore) SaveSessionState(ctx context.Context, sessionState state.SessionState) error {
	return nil
}

func (s *recordingStateStore) GetSessionState(ctx context.Context, sessionID int) (state.SessionState, error) {
	return state.SessionState{}, state.ErrStateNotFound
}

func (s *recordingStateStore) SavePartialScore(ctx context.Context, score model.ScoreResult) error {
	return nil
}

func (s *recordingStateStore) GetPartialScore(ctx context.Context, sessionID int) (model.ScoreResult, error) {
	return model.ScoreResult{}, state.ErrStateNotFound
}

func (s *recordingStateStore) AppendCorrection(ctx context.Context, correction model.CorrectionResult) error {
	return nil
}

func (s *recordingStateStore) ListCorrections(ctx context.Context, sessionID int) ([]model.CorrectionResult, error) {
	return nil, state.ErrStateNotFound
}

func (s *recordingStateStore) SaveWebSocketConnection(ctx context.Context, connection state.WebSocketConnectionState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connections[connection.SessionID] = connection
	return nil
}

func (s *recordingStateStore) GetWebSocketConnection(ctx context.Context, sessionID int) (state.WebSocketConnectionState, error) {
	connection, ok := s.connection(sessionID)
	if !ok {
		return state.WebSocketConnectionState{}, state.ErrStateNotFound
	}

	return connection, nil
}
