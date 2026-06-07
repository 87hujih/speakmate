package stream

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestRedisBusPublishesEventsThroughPubSubAndKeepsTTLList(t *testing.T) {
	ctx := context.Background()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})
	bus := NewRedisBus(client, WithRedisBusEventTTL(30*time.Minute))

	events, unsubscribe, err := bus.Subscribe(7)
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	defer unsubscribe()

	if err := bus.Publish(Event{
		Type:      EventTypeAIMessageDone,
		SessionID: 7,
		Payload: AIMessageDonePayload{
			MessageID: 11,
			Content:   "hello",
			Stage:     "项目经历",
		},
	}); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	received := receiveEvent(t, events)
	if received.Type != EventTypeAIMessageDone {
		t.Fatalf("event type = %q, want %q", received.Type, EventTypeAIMessageDone)
	}
	if received.SessionID != 7 {
		t.Fatalf("event session id = %d, want 7", received.SessionID)
	}
	if received.CreatedAt.IsZero() {
		t.Fatal("created_at is zero")
	}

	key := "session:7:events"
	if got := client.LLen(ctx, key).Val(); got != 1 {
		t.Fatalf("%s length = %d, want 1 retained event", key, got)
	}
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("TTL returned error: %v", err)
	}
	if ttl <= 0 || ttl > 30*time.Minute {
		t.Fatalf("events TTL = %s, want positive <= 30m", ttl)
	}
}
