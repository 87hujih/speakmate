package service_test

import (
	"errors"
	"testing"
	"time"

	"speakmate/internal/model"
	"speakmate/internal/repository"
	"speakmate/internal/service"
)

func TestSessionServiceCreatesRunningSessionFromScenario(t *testing.T) {
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:             1,
				Code:           "interview",
				Name:           "英语面试",
				OpeningMessage: "hello",
			},
		},
	}
	sessionRepo := newFakeSessionRepository()
	sessionService := service.NewSessionService(scenarioReader, sessionRepo)

	result, err := sessionService.CreateSession(service.CreateSessionInput{ScenarioID: 1})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	if result.Session.ID <= 0 {
		t.Fatalf("session id = %d, want positive", result.Session.ID)
	}
	if result.Session.UserID != 1 {
		t.Fatalf("user id = %d, want default 1", result.Session.UserID)
	}
	if result.Session.ScenarioID != 1 {
		t.Fatalf("scenario id = %d, want 1", result.Session.ScenarioID)
	}
	if result.Session.Status != model.SessionStatusRunning {
		t.Fatalf("status = %q, want %q", result.Session.Status, model.SessionStatusRunning)
	}
	if result.Session.TurnCount != 0 {
		t.Fatalf("turn count = %d, want 0", result.Session.TurnCount)
	}
	if len(result.Session.Messages) != 0 {
		t.Fatalf("messages length = %d, want 0", len(result.Session.Messages))
	}
	if result.Session.CreatedAt.IsZero() {
		t.Fatal("created at is zero")
	}
	if result.OpeningMessage != "hello" {
		t.Fatalf("opening message = %q, want %q", result.OpeningMessage, "hello")
	}
}

func TestSessionServiceReturnsScenarioNotFoundBeforeCreatingSession(t *testing.T) {
	scenarioReader := fakeScenarioReader{scenarios: map[int]model.Scenario{}}
	sessionRepo := newFakeSessionRepository()
	sessionService := service.NewSessionService(scenarioReader, sessionRepo)

	_, err := sessionService.CreateSession(service.CreateSessionInput{ScenarioID: 999})

	if !errors.Is(err, service.ErrScenarioNotFound) {
		t.Fatalf("error = %v, want ErrScenarioNotFound", err)
	}
	if sessionRepo.createCount != 0 {
		t.Fatalf("create count = %d, want 0", sessionRepo.createCount)
	}
}

func TestSessionServiceRejectsAlreadyFinishedSession(t *testing.T) {
	sessionRepo := newFakeSessionRepository()
	created, err := sessionRepo.Create(model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusFinished,
		CreatedAt:  time.Now(),
		Messages:   []model.Message{},
	})
	if err != nil {
		t.Fatalf("setup session returned error: %v", err)
	}
	sessionService := service.NewSessionService(fakeScenarioReader{}, sessionRepo)

	_, err = sessionService.FinishSession(created.ID)

	if !errors.Is(err, service.ErrSessionAlreadyFinished) {
		t.Fatalf("error = %v, want ErrSessionAlreadyFinished", err)
	}
}

type fakeScenarioReader struct {
	scenarios map[int]model.Scenario
}

func (r fakeScenarioReader) GetScenario(id int) (model.Scenario, error) {
	scenario, ok := r.scenarios[id]
	if !ok {
		return model.Scenario{}, service.ErrScenarioNotFound
	}

	return scenario, nil
}

type fakeSessionRepository struct {
	nextID      int
	createCount int
	sessions    map[int]model.Session
}

func newFakeSessionRepository() *fakeSessionRepository {
	return &fakeSessionRepository{
		nextID:   1,
		sessions: make(map[int]model.Session),
	}
}

func (r *fakeSessionRepository) Create(session model.Session) (model.Session, error) {
	r.createCount++
	session.ID = r.nextID
	session.SessionNo = "STEST"
	r.nextID++
	r.sessions[session.ID] = session

	return session, nil
}

func (r *fakeSessionRepository) FindByID(id int) (model.Session, error) {
	session, ok := r.sessions[id]
	if !ok {
		return model.Session{}, repository.ErrSessionNotFound
	}

	return session, nil
}

func (r *fakeSessionRepository) Finish(id int, endedAt time.Time) (model.Session, error) {
	session, ok := r.sessions[id]
	if !ok {
		return model.Session{}, repository.ErrSessionNotFound
	}
	if session.Status == model.SessionStatusFinished {
		return model.Session{}, repository.ErrSessionAlreadyFinished
	}

	session.Status = model.SessionStatusFinished
	session.EndedAt = &endedAt
	r.sessions[id] = session

	return session, nil
}
