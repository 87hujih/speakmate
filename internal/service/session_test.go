package service_test

import (
	"errors"
	"strings"
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

func TestSessionServiceSendMessageAppendsUserAndAssistantMessages(t *testing.T) {
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:   1,
				Code: "interview",
				Stages: []model.ScenarioStage{
					{Name: "自我介绍"},
					{Name: "项目经历"},
					{Name: "技术追问"},
				},
			},
		},
	}
	sessionRepo := repository.NewMemorySessionRepository()
	sessionService := service.NewSessionService(scenarioReader, sessionRepo)
	created, err := sessionService.CreateSession(service.CreateSessionInput{ScenarioID: 1})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.Session.ID,
		Content:   "  I built a robot control project last semester.  ",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if result.TurnCount != 1 {
		t.Fatalf("turn count = %d, want 1", result.TurnCount)
	}
	if result.Stage != "项目经历" {
		t.Fatalf("stage = %q, want %q", result.Stage, "项目经历")
	}
	if result.UserMessage.ID != 1 {
		t.Fatalf("user message id = %d, want 1", result.UserMessage.ID)
	}
	if result.UserMessage.SessionID != created.Session.ID {
		t.Fatalf("user message session id = %d, want %d", result.UserMessage.SessionID, created.Session.ID)
	}
	if result.UserMessage.Role != model.MessageRoleUser {
		t.Fatalf("user message role = %q, want %q", result.UserMessage.Role, model.MessageRoleUser)
	}
	if result.UserMessage.Content != "I built a robot control project last semester." {
		t.Fatalf("user message content = %q, want trimmed content", result.UserMessage.Content)
	}
	if result.UserMessage.Stage != "自我介绍" {
		t.Fatalf("user message stage = %q, want %q", result.UserMessage.Stage, "自我介绍")
	}
	if result.UserMessage.CreatedAt.IsZero() {
		t.Fatal("user message created_at is zero")
	}
	if result.AIMessage.ID != 2 {
		t.Fatalf("ai message id = %d, want 2", result.AIMessage.ID)
	}
	if result.AIMessage.Role != model.MessageRoleAssistant {
		t.Fatalf("ai message role = %q, want %q", result.AIMessage.Role, model.MessageRoleAssistant)
	}
	if result.AIMessage.Stage != "项目经历" {
		t.Fatalf("ai message stage = %q, want %q", result.AIMessage.Stage, "项目经历")
	}
	if !strings.Contains(strings.ToLower(result.AIMessage.Content), "project") {
		t.Fatalf("ai message content = %q, want interview project-related reply", result.AIMessage.Content)
	}

	detail, err := sessionService.GetSession(created.Session.ID)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if detail.Session.TurnCount != 1 {
		t.Fatalf("persisted turn count = %d, want 1", detail.Session.TurnCount)
	}
	if len(detail.Session.Messages) != 2 {
		t.Fatalf("persisted message count = %d, want 2", len(detail.Session.Messages))
	}
	if detail.Session.Messages[0].Content != result.UserMessage.Content {
		t.Fatalf("persisted user content = %q, want %q", detail.Session.Messages[0].Content, result.UserMessage.Content)
	}
	if detail.Session.Messages[1].Content != result.AIMessage.Content {
		t.Fatalf("persisted ai content = %q, want %q", detail.Session.Messages[1].Content, result.AIMessage.Content)
	}
}

func TestSessionServiceSendMessageAdvancesTurnCount(t *testing.T) {
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:   1,
				Code: "meeting",
				Stages: []model.ScenarioStage{
					{Name: "进度同步"},
					{Name: "观点表达"},
					{Name: "澄清确认"},
				},
			},
		},
	}
	sessionRepo := repository.NewMemorySessionRepository()
	sessionService := service.NewSessionService(scenarioReader, sessionRepo)
	created, err := sessionService.CreateSession(service.CreateSessionInput{ScenarioID: 1})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	first, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.Session.ID,
		Content:   "The API work is mostly finished.",
	})
	if err != nil {
		t.Fatalf("first SendMessage returned error: %v", err)
	}
	second, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.Session.ID,
		Content:   "I think we should prioritize the report page.",
	})
	if err != nil {
		t.Fatalf("second SendMessage returned error: %v", err)
	}

	if first.TurnCount != 1 {
		t.Fatalf("first turn count = %d, want 1", first.TurnCount)
	}
	if second.TurnCount != 2 {
		t.Fatalf("second turn count = %d, want 2", second.TurnCount)
	}
	if second.UserMessage.ID != 3 || second.AIMessage.ID != 4 {
		t.Fatalf("second message ids = %d/%d, want 3/4", second.UserMessage.ID, second.AIMessage.ID)
	}
	if second.Stage != "澄清确认" {
		t.Fatalf("second stage = %q, want %q", second.Stage, "澄清确认")
	}
}

func TestSessionServiceSendMessageRejectsEmptyContent(t *testing.T) {
	sessionService := service.NewSessionService(fakeScenarioReader{}, repository.NewMemorySessionRepository())

	_, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: 1,
		Content:   "   ",
	})

	if !errors.Is(err, service.ErrMessageContentRequired) {
		t.Fatalf("error = %v, want ErrMessageContentRequired", err)
	}
}

func TestSessionServiceSendMessageRejectsFinishedSession(t *testing.T) {
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:   1,
				Code: "restaurant",
				Stages: []model.ScenarioStage{
					{Name: "询问菜单"},
					{Name: "表达偏好"},
					{Name: "确认订单"},
				},
			},
		},
	}
	sessionRepo := repository.NewMemorySessionRepository()
	sessionService := service.NewSessionService(scenarioReader, sessionRepo)
	created, err := sessionService.CreateSession(service.CreateSessionInput{ScenarioID: 1})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	if _, err := sessionService.FinishSession(created.Session.ID); err != nil {
		t.Fatalf("FinishSession returned error: %v", err)
	}

	_, err = sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.Session.ID,
		Content:   "Can I have the soup?",
	})

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

func (r *fakeSessionRepository) AddMessageTurn(id int, build func(model.Session, int, int) (model.Message, model.Message, error)) (model.Session, error) {
	session, ok := r.sessions[id]
	if !ok {
		return model.Session{}, repository.ErrSessionNotFound
	}
	if session.Status == model.SessionStatusFinished {
		return model.Session{}, repository.ErrSessionAlreadyFinished
	}

	nextID := len(session.Messages) + 1
	userMessage, aiMessage, err := build(session, nextID, nextID+1)
	if err != nil {
		return model.Session{}, err
	}
	session.Messages = append(session.Messages, userMessage, aiMessage)
	session.TurnCount++
	r.sessions[id] = session

	return session, nil
}
