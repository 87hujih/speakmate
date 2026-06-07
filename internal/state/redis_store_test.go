package state

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"speakmate/internal/model"
)

func TestRedisSessionStateStoreUsesDocumentedKeysAndTTLs(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})
	store := NewRedisSessionStateStore(
		client,
		WithRedisStateTTL(2*time.Hour),
		WithRedisWebSocketTTL(30*time.Minute),
	)
	now := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)

	if err := store.SaveMessageSnapshot(ctx, 7, []model.Message{
		{ID: 1, SessionID: 7, Role: model.MessageRoleUser, Content: "hello", CreatedAt: now},
		{ID: 2, SessionID: 7, Role: model.MessageRoleAI, Content: "hi", CreatedAt: now},
	}); err != nil {
		t.Fatalf("SaveMessageSnapshot returned error: %v", err)
	}
	if err := store.SaveSessionState(ctx, SessionState{
		SessionID:  7,
		ScenarioID: 1,
		UserID:     42,
		Status:     "running",
		Stage:      "项目经历",
		TurnCount:  1,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("SaveSessionState returned error: %v", err)
	}
	if err := store.SavePartialScore(ctx, model.ScoreResult{
		MessageID: 1, SessionID: 7, Fluency: 75, Grammar: 72, Expression: 80, Vocabulary: 76, Completion: 85, TotalScore: 77, Comment: "stable",
	}); err != nil {
		t.Fatalf("SavePartialScore returned error: %v", err)
	}
	if err := store.AppendCorrection(ctx, model.CorrectionResult{
		MessageID: 1, SessionID: 7, OriginalText: "bad", CorrectedText: "good",
	}); err != nil {
		t.Fatalf("AppendCorrection returned error: %v", err)
	}
	if err := store.SaveWebSocketConnection(ctx, WebSocketConnectionState{
		SessionID: 7, Status: "receiving", ContentType: "audio/ogg", ChunkCount: 3, LastSequence: 3, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("SaveWebSocketConnection returned error: %v", err)
	}

	assertRedisTTL(t, ctx, client, "session:7:messages", 2*time.Hour)
	assertRedisTTL(t, ctx, client, "session:7:state", 2*time.Hour)
	assertRedisTTL(t, ctx, client, "session:7:partial_score", 2*time.Hour)
	assertRedisTTL(t, ctx, client, "session:7:corrections", 2*time.Hour)
	assertRedisTTL(t, ctx, client, "ws:7:connection", 30*time.Minute)
	if got := client.LLen(ctx, "session:7:messages").Val(); got != 2 {
		t.Fatalf("session:7:messages length = %d, want 2", got)
	}
	if got := client.LLen(ctx, "session:7:corrections").Val(); got != 1 {
		t.Fatalf("session:7:corrections length = %d, want 1", got)
	}

	state, err := store.GetSessionState(ctx, 7)
	if err != nil {
		t.Fatalf("GetSessionState returned error: %v", err)
	}
	if state.Stage != "项目经历" || state.TurnCount != 1 {
		t.Fatalf("state = %+v, want saved state", state)
	}
	score, err := store.GetPartialScore(ctx, 7)
	if err != nil {
		t.Fatalf("GetPartialScore returned error: %v", err)
	}
	if score.TotalScore != 77 || score.Expression != 80 {
		t.Fatalf("score = %+v, want saved score", score)
	}
	connection, err := store.GetWebSocketConnection(ctx, 7)
	if err != nil {
		t.Fatalf("GetWebSocketConnection returned error: %v", err)
	}
	if connection.Status != "receiving" || connection.LastSequence != 3 {
		t.Fatalf("connection = %+v, want saved connection state", connection)
	}
}

func assertRedisTTL(t *testing.T, ctx context.Context, client *goredis.Client, key string, want time.Duration) {
	t.Helper()

	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("TTL(%s) returned error: %v", key, err)
	}
	if ttl <= 0 {
		t.Fatalf("TTL(%s) = %s, want positive TTL", key, ttl)
	}
	if ttl > want {
		t.Fatalf("TTL(%s) = %s, want <= %s", key, ttl, want)
	}
}
