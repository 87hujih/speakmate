package stream

import (
	"errors"
	"testing"
	"time"
)

func TestBusPublishesEventsToSessionSubscribers(t *testing.T) {
	bus := NewBus()
	events, unsubscribe, err := bus.Subscribe(7)
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	defer unsubscribe()

	published := Event{
		Type:      EventTypeAIMessageDone,
		SessionID: 7,
		Payload: map[string]string{
			"content": "hello",
		},
	}
	if err := bus.Publish(published); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	received := receiveEvent(t, events)
	if received.Type != EventTypeAIMessageDone {
		t.Fatalf("event type = %q, want %q", received.Type, EventTypeAIMessageDone)
	}
	if received.SessionID != 7 {
		t.Fatalf("session id = %d, want 7", received.SessionID)
	}
	if received.CreatedAt.IsZero() {
		t.Fatal("created_at is zero")
	}
}

func TestBusDoesNotDeliverEventsAcrossSessions(t *testing.T) {
	bus := NewBus()
	events, unsubscribe, err := bus.Subscribe(8)
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	defer unsubscribe()

	if err := bus.Publish(Event{Type: EventTypeScoreUpdated, SessionID: 7}); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	select {
	case event := <-events:
		t.Fatalf("received event for wrong session: %+v", event)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestBusUnsubscribeClosesSubscription(t *testing.T) {
	bus := NewBus()
	events, unsubscribe, err := bus.Subscribe(7)
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}

	unsubscribe()

	if _, ok := <-events; ok {
		t.Fatal("subscription channel still open after unsubscribe")
	}
	if err := bus.Publish(Event{Type: EventTypeAIMessageDone, SessionID: 7}); err != nil {
		t.Fatalf("Publish returned error after unsubscribe: %v", err)
	}
}

func TestBusCloseClosesSubscribersAndRejectsNewWork(t *testing.T) {
	bus := NewBus()
	events, unsubscribe, err := bus.Subscribe(7)
	if err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	defer unsubscribe()

	bus.Close()

	if _, ok := <-events; ok {
		t.Fatal("subscription channel still open after bus close")
	}
	if err := bus.Publish(Event{Type: EventTypeError, SessionID: 7}); !errors.Is(err, ErrBusClosed) {
		t.Fatalf("Publish error = %v, want ErrBusClosed", err)
	}
	if _, _, err := bus.Subscribe(7); !errors.Is(err, ErrBusClosed) {
		t.Fatalf("Subscribe error = %v, want ErrBusClosed", err)
	}
}

func receiveEvent(t *testing.T, events <-chan Event) Event {
	t.Helper()

	select {
	case event := <-events:
		return event
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for event")
		return Event{}
	}
}
