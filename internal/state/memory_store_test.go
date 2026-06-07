package state

import (
	"context"
	"errors"
	"testing"
	"time"

	"speakmate/internal/model"
)

func TestMemorySessionStateStoreSavesShortLivedSessionState(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)
	store := NewMemorySessionStateStore(WithMemoryClock(func() time.Time {
		return now
	}))

	messages := []model.Message{
		{ID: 1, SessionID: 7, Role: model.MessageRoleUser, Content: "I am study computer science.", Stage: "自我介绍", CreatedAt: now},
		{ID: 2, SessionID: 7, Role: model.MessageRoleAI, Content: "Could you describe a project?", Stage: "项目经历", CreatedAt: now},
	}
	if err := store.SaveMessageSnapshot(ctx, 7, messages); err != nil {
		t.Fatalf("SaveMessageSnapshot returned error: %v", err)
	}
	if err := store.SaveSessionState(ctx, SessionState{
		SessionID:  7,
		ScenarioID: 1,
		UserID:     42,
		Status:     string(model.SessionStatusRunning),
		Stage:      "项目经历",
		TurnCount:  1,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("SaveSessionState returned error: %v", err)
	}
	score := model.ScoreResult{
		MessageID:  1,
		SessionID:  7,
		Fluency:    75,
		Grammar:    72,
		Expression: 80,
		Vocabulary: 76,
		Completion: 85,
		TotalScore: 77,
		Comment:    "stable score",
	}
	if err := store.SavePartialScore(ctx, score); err != nil {
		t.Fatalf("SavePartialScore returned error: %v", err)
	}
	correction := model.CorrectionResult{
		MessageID:     1,
		SessionID:     7,
		OriginalText:  "I am study computer science.",
		CorrectedText: "I am studying computer science.",
		Errors: []model.CorrectionError{
			{Type: model.CorrectionErrorTypeGrammar, Span: "am study", Suggestion: "am studying"},
		},
	}
	if err := store.AppendCorrection(ctx, correction); err != nil {
		t.Fatalf("AppendCorrection returned error: %v", err)
	}
	if err := store.SaveWebSocketConnection(ctx, WebSocketConnectionState{
		SessionID:    7,
		Status:       "receiving",
		ContentType:  "audio/ogg",
		ChunkCount:   2,
		LastSequence: 2,
		UpdatedAt:    now,
	}); err != nil {
		t.Fatalf("SaveWebSocketConnection returned error: %v", err)
	}

	gotMessages, err := store.GetMessageSnapshot(ctx, 7)
	if err != nil {
		t.Fatalf("GetMessageSnapshot returned error: %v", err)
	}
	if len(gotMessages) != 2 || gotMessages[1].Content != "Could you describe a project?" {
		t.Fatalf("message snapshot = %+v, want saved messages", gotMessages)
	}
	gotState, err := store.GetSessionState(ctx, 7)
	if err != nil {
		t.Fatalf("GetSessionState returned error: %v", err)
	}
	if gotState.Stage != "项目经历" || gotState.TurnCount != 1 || gotState.Status != "running" {
		t.Fatalf("session state = %+v, want current stage/turn/status", gotState)
	}
	gotScore, err := store.GetPartialScore(ctx, 7)
	if err != nil {
		t.Fatalf("GetPartialScore returned error: %v", err)
	}
	if gotScore.TotalScore != 77 || gotScore.Grammar != 72 {
		t.Fatalf("partial score = %+v, want saved score", gotScore)
	}
	gotCorrections, err := store.ListCorrections(ctx, 7)
	if err != nil {
		t.Fatalf("ListCorrections returned error: %v", err)
	}
	if len(gotCorrections) != 1 || gotCorrections[0].CorrectedText != "I am studying computer science." {
		t.Fatalf("corrections = %+v, want saved correction", gotCorrections)
	}
	gotConnection, err := store.GetWebSocketConnection(ctx, 7)
	if err != nil {
		t.Fatalf("GetWebSocketConnection returned error: %v", err)
	}
	if gotConnection.Status != "receiving" || gotConnection.ChunkCount != 2 {
		t.Fatalf("connection = %+v, want receiving state", gotConnection)
	}
}

func TestMemorySessionStateStoreExpiresStateByTTL(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)
	store := NewMemorySessionStateStore(
		WithMemoryClock(func() time.Time { return now }),
		WithMemoryStateTTL(2*time.Second),
		WithMemoryWebSocketTTL(time.Second),
	)

	if err := store.SaveSessionState(ctx, SessionState{SessionID: 7, Status: "running", UpdatedAt: now}); err != nil {
		t.Fatalf("SaveSessionState returned error: %v", err)
	}
	if err := store.SaveWebSocketConnection(ctx, WebSocketConnectionState{SessionID: 7, Status: "started", UpdatedAt: now}); err != nil {
		t.Fatalf("SaveWebSocketConnection returned error: %v", err)
	}

	now = now.Add(1500 * time.Millisecond)
	if _, err := store.GetWebSocketConnection(ctx, 7); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("GetWebSocketConnection error = %v, want ErrStateNotFound after websocket TTL", err)
	}
	if _, err := store.GetSessionState(ctx, 7); err != nil {
		t.Fatalf("GetSessionState returned error before state TTL: %v", err)
	}

	now = now.Add(time.Second)
	if _, err := store.GetSessionState(ctx, 7); !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("GetSessionState error = %v, want ErrStateNotFound after state TTL", err)
	}
}
