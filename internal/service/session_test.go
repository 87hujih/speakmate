package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"speakmate/internal/agent"
	"speakmate/internal/model"
	"speakmate/internal/repository"
	"speakmate/internal/service"
	"speakmate/internal/state"
	"speakmate/internal/stream"
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

func TestSessionServiceSendsMessageAndPersistsMockReply(t *testing.T) {
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:   1,
				Code: "interview",
				Stages: []model.ScenarioStage{
					{Name: "自我介绍"},
					{Name: "项目经历"},
				},
			},
		},
	}
	sessionRepo := newFakeSessionRepository()
	created, err := sessionRepo.Create(model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusRunning,
		CreatedAt:  time.Now(),
		Messages:   []model.Message{},
	})
	if err != nil {
		t.Fatalf("setup session returned error: %v", err)
	}
	sessionService := service.NewSessionService(scenarioReader, sessionRepo)

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   " I built a robot control project. ",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if result.UserMessage.Role != model.MessageRoleUser {
		t.Fatalf("user role = %q, want %q", result.UserMessage.Role, model.MessageRoleUser)
	}
	if result.UserMessage.Content != "I built a robot control project." {
		t.Fatalf("user content = %q, want trimmed content", result.UserMessage.Content)
	}
	if result.AIMessage.Role != model.MessageRoleAI {
		t.Fatalf("ai role = %q, want %q", result.AIMessage.Role, model.MessageRoleAI)
	}
	if result.AIMessage.Content == "" {
		t.Fatal("ai content is empty")
	}
	if result.Stage != "项目经历" {
		t.Fatalf("stage = %q, want 项目经历", result.Stage)
	}
	if result.TurnCount != 1 {
		t.Fatalf("turn_count = %d, want 1", result.TurnCount)
	}
	if result.NextGoal == "" {
		t.Fatal("next_goal is empty")
	}

	saved, err := sessionRepo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if saved.TurnCount != 1 {
		t.Fatalf("saved turn_count = %d, want 1", saved.TurnCount)
	}
	if len(saved.Messages) != 2 {
		t.Fatalf("saved messages length = %d, want 2", len(saved.Messages))
	}
}

func TestSessionServiceSendsMessageAndGeneratesFeedback(t *testing.T) {
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:   1,
				Code: "interview",
				Stages: []model.ScenarioStage{
					{Name: "自我介绍"},
					{Name: "项目经历"},
				},
			},
		},
	}
	sessionRepo := newFakeSessionRepository()
	created, err := sessionRepo.Create(model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusRunning,
		CreatedAt:  time.Now(),
		Messages:   []model.Message{},
	})
	if err != nil {
		t.Fatalf("setup session returned error: %v", err)
	}
	feedbackRepo := newFakeFeedbackRepository()
	correctionAgent := &fakeCorrectionAgent{}
	scoringAgent := &fakeScoringAgent{}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(correctionAgent),
		service.WithScoringAgent(scoringAgent),
	)

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I am study computer science and I have did a project.",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if correctionAgent.callCount != 1 {
		t.Fatalf("correction call count = %d, want 1", correctionAgent.callCount)
	}
	if scoringAgent.callCount != 1 {
		t.Fatalf("scoring call count = %d, want 1", scoringAgent.callCount)
	}
	if correctionAgent.input.UserMessage.ID != result.UserMessage.ID {
		t.Fatalf("correction message id = %d, want saved user message id %d", correctionAgent.input.UserMessage.ID, result.UserMessage.ID)
	}
	if scoringAgent.input.Correction.MessageID != result.UserMessage.ID {
		t.Fatalf("scoring correction message id = %d, want %d", scoringAgent.input.Correction.MessageID, result.UserMessage.ID)
	}

	correction, ok := feedbackRepo.correctionsByMessageID[result.UserMessage.ID]
	if !ok {
		t.Fatalf("saved correction for message %d not found", result.UserMessage.ID)
	}
	if correction.SessionID != created.ID {
		t.Fatalf("correction session id = %d, want %d", correction.SessionID, created.ID)
	}
	if len(correction.Errors) != 2 {
		t.Fatalf("correction errors length = %d, want 2", len(correction.Errors))
	}

	score, ok := feedbackRepo.scoresBySessionID[created.ID]
	if !ok {
		t.Fatalf("saved score for session %d not found", created.ID)
	}
	if score.MessageID != result.UserMessage.ID {
		t.Fatalf("score message id = %d, want %d", score.MessageID, result.UserMessage.ID)
	}
	if score.TotalScore != 77 {
		t.Fatalf("score total = %d, want 77", score.TotalScore)
	}

	if !result.CorrectionSummary.HasErrors {
		t.Fatal("correction summary has_errors = false, want true")
	}
	if result.CorrectionSummary.ErrorCount != 2 {
		t.Fatalf("correction summary error_count = %d, want 2", result.CorrectionSummary.ErrorCount)
	}
	if result.ScoreSummary.TotalScore != 77 {
		t.Fatalf("score summary total_score = %d, want 77", result.ScoreSummary.TotalScore)
	}
	if result.ScoreSummary.Grammar != 72 {
		t.Fatalf("score summary grammar = %d, want 72", result.ScoreSummary.Grammar)
	}
	if result.ScoreSummary.Expression != 80 {
		t.Fatalf("score summary expression = %d, want 80", result.ScoreSummary.Expression)
	}
}

func TestSessionServicePublishesMessageFeedbackStreamEvents(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	correctionAgent := &fakeCorrectionAgent{}
	scoringAgent := &fakeScoringAgent{}
	publisher := &fakeEventPublisher{}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(&fakeConversationAgent{
			output: agent.ConversationOutput{
				Reply:    "What was your role?",
				Stage:    "项目经历",
				NextGoal: "ask user to describe personal project contribution",
			},
		}),
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(correctionAgent),
		service.WithScoringAgent(scoringAgent),
		service.WithEventPublisher(publisher),
	)

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I am study computer science and I have did a project.",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	wantTypes := []stream.EventType{
		stream.EventTypeAIMessageDelta,
		stream.EventTypeAIMessageDone,
		stream.EventTypeCorrectionDone,
		stream.EventTypeScoreUpdated,
	}
	if len(publisher.events) != len(wantTypes) {
		t.Fatalf("published events length = %d, want %d: %+v", len(publisher.events), len(wantTypes), publisher.events)
	}
	for i, wantType := range wantTypes {
		if publisher.events[i].Type != wantType {
			t.Fatalf("event[%d] type = %q, want %q", i, publisher.events[i].Type, wantType)
		}
		if publisher.events[i].SessionID != created.ID {
			t.Fatalf("event[%d] session id = %d, want %d", i, publisher.events[i].SessionID, created.ID)
		}
	}

	delta, ok := publisher.events[0].Payload.(stream.AIMessageDeltaPayload)
	if !ok {
		t.Fatalf("delta payload type = %T, want AIMessageDeltaPayload", publisher.events[0].Payload)
	}
	if delta.MessageID != result.AIMessage.ID {
		t.Fatalf("delta message id = %d, want %d", delta.MessageID, result.AIMessage.ID)
	}
	if delta.Delta != result.AIMessage.Content {
		t.Fatalf("delta = %q, want full simulated reply chunk %q", delta.Delta, result.AIMessage.Content)
	}

	done, ok := publisher.events[1].Payload.(stream.AIMessageDonePayload)
	if !ok {
		t.Fatalf("done payload type = %T, want AIMessageDonePayload", publisher.events[1].Payload)
	}
	if done.MessageID != result.AIMessage.ID || done.Content != result.AIMessage.Content {
		t.Fatalf("done payload = %+v, want saved ai message %d/%q", done, result.AIMessage.ID, result.AIMessage.Content)
	}

	correction, ok := publisher.events[2].Payload.(stream.CorrectionDonePayload)
	if !ok {
		t.Fatalf("correction payload type = %T, want CorrectionDonePayload", publisher.events[2].Payload)
	}
	if correction.MessageID != result.UserMessage.ID {
		t.Fatalf("correction message id = %d, want %d", correction.MessageID, result.UserMessage.ID)
	}
	if !correction.HasErrors || correction.ErrorCount != 2 {
		t.Fatalf("correction payload = %+v, want has_errors with 2 errors", correction)
	}

	score, ok := publisher.events[3].Payload.(stream.ScoreUpdatedPayload)
	if !ok {
		t.Fatalf("score payload type = %T, want ScoreUpdatedPayload", publisher.events[3].Payload)
	}
	if score.MessageID != result.UserMessage.ID {
		t.Fatalf("score message id = %d, want %d", score.MessageID, result.UserMessage.ID)
	}
	if score.TotalScore != 77 || score.Grammar != 72 || score.Expression != 80 {
		t.Fatalf("score payload = %+v, want saved score summary", score)
	}
}

func TestSessionServiceWritesShortLivedStateDuringLifecycle(t *testing.T) {
	scenarioReader, sessionRepo, _ := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	stateStore := newFakeStateStore()
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(&fakeConversationAgent{
			output: agent.ConversationOutput{
				Reply:    "What was your role?",
				Stage:    "项目经历",
				NextGoal: "ask user to describe personal project contribution",
			},
		}),
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(&fakeCorrectionAgent{}),
		service.WithScoringAgent(&fakeScoringAgent{}),
		service.WithStateStore(stateStore),
	)

	created, err := sessionService.CreateSession(service.CreateSessionInput{ScenarioID: 1, UserID: 42})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	createdState := stateStore.states[created.Session.ID]
	if createdState.Status != "running" || createdState.TurnCount != 0 || createdState.UserID != 42 {
		t.Fatalf("created state = %+v, want running turn 0 user 42", createdState)
	}

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.Session.ID,
		Content:   "I am study computer science and I have did a project.",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if len(stateStore.messages[created.Session.ID]) != 2 {
		t.Fatalf("message snapshot length = %d, want 2", len(stateStore.messages[created.Session.ID]))
	}
	sentState := stateStore.states[created.Session.ID]
	if sentState.Stage != result.Stage || sentState.TurnCount != 1 || sentState.Status != "running" {
		t.Fatalf("sent state = %+v, want current stage/turn/status", sentState)
	}
	if stateStore.scores[created.Session.ID].TotalScore != 77 {
		t.Fatalf("partial score = %+v, want generated score", stateStore.scores[created.Session.ID])
	}
	if len(stateStore.corrections[created.Session.ID]) != 1 {
		t.Fatalf("corrections length = %d, want 1", len(stateStore.corrections[created.Session.ID]))
	}

	finished, err := sessionService.FinishSession(created.Session.ID)
	if err != nil {
		t.Fatalf("FinishSession returned error: %v", err)
	}
	finishedState := stateStore.states[created.Session.ID]
	if finishedState.Status != string(finished.Status) || finishedState.TurnCount != finished.TurnCount {
		t.Fatalf("finished state = %+v, want finished session state", finishedState)
	}
}

func TestSessionServiceReturnsEventPublishFailure(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	publisher := &fakeEventPublisher{err: errors.New("redis publish failed")}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(&fakeConversationAgent{
			output: agent.ConversationOutput{
				Reply: "What was your role?",
				Stage: "项目经历",
			},
		}),
		service.WithEventPublisher(publisher),
	)

	_, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I built a robot control project.",
	})

	if !errors.Is(err, service.ErrEventPublishFailed) {
		t.Fatalf("error = %v, want ErrEventPublishFailed", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events length = %d, want first failed event only", len(publisher.events))
	}
}

func TestSessionServicePublishesStreamingDeltasBeforePersistedDone(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	publisher := &fakeEventPublisher{}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(&fakeStreamingConversationAgent{
			deltas: []string{"What ", "was your role?"},
			output: agent.ConversationOutput{
				Reply:    "What was your role?",
				Stage:    "项目经历",
				NextGoal: "ask user to describe personal project contribution",
			},
		}),
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(&fakeCorrectionAgent{}),
		service.WithScoringAgent(&fakeScoringAgent{}),
		service.WithEventPublisher(publisher),
	)

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I built a robot control project.",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	wantTypes := []stream.EventType{
		stream.EventTypeAIMessageDelta,
		stream.EventTypeAIMessageDelta,
		stream.EventTypeAIMessageDone,
		stream.EventTypeCorrectionDone,
		stream.EventTypeScoreUpdated,
	}
	if len(publisher.events) != len(wantTypes) {
		t.Fatalf("published events length = %d, want %d: %+v", len(publisher.events), len(wantTypes), publisher.events)
	}
	for i, wantType := range wantTypes {
		if publisher.events[i].Type != wantType {
			t.Fatalf("event[%d] type = %q, want %q", i, publisher.events[i].Type, wantType)
		}
	}

	firstDelta := publisher.events[0].Payload.(stream.AIMessageDeltaPayload)
	if firstDelta.MessageID != 0 || firstDelta.Delta != "What " {
		t.Fatalf("first delta = %+v, want unsaved message id 0 and first chunk", firstDelta)
	}
	secondDelta := publisher.events[1].Payload.(stream.AIMessageDeltaPayload)
	if secondDelta.MessageID != 0 || secondDelta.Delta != "was your role?" {
		t.Fatalf("second delta = %+v, want unsaved message id 0 and second chunk", secondDelta)
	}

	done := publisher.events[2].Payload.(stream.AIMessageDonePayload)
	if done.MessageID != result.AIMessage.ID {
		t.Fatalf("done message id = %d, want saved ai message id %d", done.MessageID, result.AIMessage.ID)
	}
	if done.Content != "What was your role?" {
		t.Fatalf("done content = %q, want complete streamed reply", done.Content)
	}

	saved, err := sessionRepo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if len(saved.Messages) != 2 {
		t.Fatalf("saved messages length = %d, want 2", len(saved.Messages))
	}
	if saved.Messages[1].Content != "What was your role?" {
		t.Fatalf("saved ai content = %q, want complete streamed reply", saved.Messages[1].Content)
	}
}

func TestSessionServicePublishesErrorWhenStreamingConversationFails(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	publisher := &fakeEventPublisher{}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(&fakeStreamingConversationAgent{err: errors.New("stream failed")}),
		service.WithEventPublisher(publisher),
	)

	_, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "Hello",
	})

	if !errors.Is(err, service.ErrConversationAgentFailed) {
		t.Fatalf("error = %v, want ErrConversationAgentFailed", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events length = %d, want 1: %+v", len(publisher.events), publisher.events)
	}
	event := publisher.events[0]
	if event.Type != stream.EventTypeError {
		t.Fatalf("event type = %q, want %q", event.Type, stream.EventTypeError)
	}
	payload, ok := event.Payload.(stream.ErrorPayload)
	if !ok {
		t.Fatalf("payload type = %T, want ErrorPayload", event.Payload)
	}
	if payload.Code != "conversation_agent_failed" || payload.Message != "对话 AI 回复失败" {
		t.Fatalf("payload = %+v, want conversation failure", payload)
	}

	saved, findErr := sessionRepo.FindByID(created.ID)
	if findErr != nil {
		t.Fatalf("FindByID returned error: %v", findErr)
	}
	if len(saved.Messages) != 0 {
		t.Fatalf("saved messages length = %d, want 0", len(saved.Messages))
	}
}

func TestSessionServiceFailOpenKeepsMessageChainWhenCorrectionAgentFails(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	correctionAgent := &fakeCorrectionAgent{err: errors.New("fake correction failed")}
	scoringAgent := &fakeScoringAgent{}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(correctionAgent),
		service.WithScoringAgent(scoringAgent),
	)

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I am study computer science.",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if result.UserMessage.ID == 0 {
		t.Fatal("user message id = 0, want saved user message")
	}
	if result.AIMessage.ID == 0 {
		t.Fatal("ai message id = 0, want saved ai message")
	}
	if result.CorrectionSummary != (service.CorrectionSummary{}) {
		t.Fatalf("correction summary = %+v, want empty summary", result.CorrectionSummary)
	}
	if result.ScoreSummary != (service.ScoreSummary{}) {
		t.Fatalf("score summary = %+v, want empty summary", result.ScoreSummary)
	}
	if correctionAgent.callCount != 1 {
		t.Fatalf("correction call count = %d, want 1", correctionAgent.callCount)
	}
	if scoringAgent.callCount != 0 {
		t.Fatalf("scoring call count = %d, want 0", scoringAgent.callCount)
	}
	if feedbackRepo.callCount != 0 {
		t.Fatalf("feedback repo call count = %d, want 0", feedbackRepo.callCount)
	}

	saved, findErr := sessionRepo.FindByID(created.ID)
	if findErr != nil {
		t.Fatalf("FindByID returned error: %v", findErr)
	}
	if len(saved.Messages) != 2 {
		t.Fatalf("saved messages length = %d, want 2", len(saved.Messages))
	}
}

func TestSessionServiceFailOpenKeepsCorrectionSummaryWhenScoringAgentFails(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	correctionAgent := &fakeCorrectionAgent{}
	scoringAgent := &fakeScoringAgent{err: errors.New("fake scoring failed")}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(correctionAgent),
		service.WithScoringAgent(scoringAgent),
	)

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I am study computer science and I have did a project.",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if result.CorrectionSummary.ErrorCount != 2 {
		t.Fatalf("correction summary error_count = %d, want 2", result.CorrectionSummary.ErrorCount)
	}
	if result.ScoreSummary != (service.ScoreSummary{}) {
		t.Fatalf("score summary = %+v, want empty summary", result.ScoreSummary)
	}
	if correctionAgent.callCount != 1 {
		t.Fatalf("correction call count = %d, want 1", correctionAgent.callCount)
	}
	if scoringAgent.callCount != 1 {
		t.Fatalf("scoring call count = %d, want 1", scoringAgent.callCount)
	}
	if _, ok := feedbackRepo.correctionsByMessageID[result.UserMessage.ID]; !ok {
		t.Fatalf("saved correction for message %d not found", result.UserMessage.ID)
	}
	if _, ok := feedbackRepo.scoresBySessionID[created.ID]; ok {
		t.Fatalf("saved score for session %d found, want none", created.ID)
	}
}

func TestSessionServiceFailClosedReturnsFeedbackAgentFailureWhenCorrectionAgentFails(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	correctionAgent := &fakeCorrectionAgent{err: errors.New("fake correction failed")}
	scoringAgent := &fakeScoringAgent{}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(correctionAgent),
		service.WithScoringAgent(scoringAgent),
		service.WithFeedbackFailOpen(false),
	)

	_, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I am study computer science.",
	})

	if !errors.Is(err, service.ErrFeedbackAgentFailed) {
		t.Fatalf("error = %v, want ErrFeedbackAgentFailed", err)
	}
	if scoringAgent.callCount != 0 {
		t.Fatalf("scoring call count = %d, want 0", scoringAgent.callCount)
	}
	saved, findErr := sessionRepo.FindByID(created.ID)
	if findErr != nil {
		t.Fatalf("FindByID returned error: %v", findErr)
	}
	if len(saved.Messages) != 2 {
		t.Fatalf("saved messages length = %d, want 2", len(saved.Messages))
	}
}

func TestSessionServicePublishesErrorEventWhenFeedbackFails(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	publisher := &fakeEventPublisher{}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(&fakeConversationAgent{
			output: agent.ConversationOutput{
				Reply:    "What was your role?",
				Stage:    "项目经历",
				NextGoal: "ask user to describe personal project contribution",
			},
		}),
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(&fakeCorrectionAgent{err: errors.New("fake correction failed")}),
		service.WithScoringAgent(&fakeScoringAgent{}),
		service.WithFeedbackFailOpen(false),
		service.WithEventPublisher(publisher),
	)

	_, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I am study computer science.",
	})

	if !errors.Is(err, service.ErrFeedbackAgentFailed) {
		t.Fatalf("error = %v, want ErrFeedbackAgentFailed", err)
	}
	if len(publisher.events) != 3 {
		t.Fatalf("published events length = %d, want 3: %+v", len(publisher.events), publisher.events)
	}
	event := publisher.events[2]
	if event.Type != stream.EventTypeError {
		t.Fatalf("event type = %q, want %q", event.Type, stream.EventTypeError)
	}
	if event.SessionID != created.ID {
		t.Fatalf("event session id = %d, want %d", event.SessionID, created.ID)
	}
	payload, ok := event.Payload.(stream.ErrorPayload)
	if !ok {
		t.Fatalf("payload type = %T, want ErrorPayload", event.Payload)
	}
	if payload.Code != "feedback_agent_failed" || payload.Message != "反馈 AI 生成失败" {
		t.Fatalf("payload = %+v, want feedback failure", payload)
	}
}

func TestSessionServiceFailClosedReturnsFeedbackAgentFailureWhenScoringAgentFails(t *testing.T) {
	scenarioReader, sessionRepo, created := setupFeedbackSession(t)
	feedbackRepo := newFakeFeedbackRepository()
	correctionAgent := &fakeCorrectionAgent{}
	scoringAgent := &fakeScoringAgent{err: errors.New("fake scoring failed")}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(correctionAgent),
		service.WithScoringAgent(scoringAgent),
		service.WithFeedbackFailOpen(false),
	)

	_, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "I am study computer science and I have did a project.",
	})

	if !errors.Is(err, service.ErrFeedbackAgentFailed) {
		t.Fatalf("error = %v, want ErrFeedbackAgentFailed", err)
	}
	if correctionAgent.callCount != 1 {
		t.Fatalf("correction call count = %d, want 1", correctionAgent.callCount)
	}
	if scoringAgent.callCount != 1 {
		t.Fatalf("scoring call count = %d, want 1", scoringAgent.callCount)
	}
	if len(feedbackRepo.correctionsByMessageID) != 1 {
		t.Fatalf("saved corrections length = %d, want 1", len(feedbackRepo.correctionsByMessageID))
	}
	if len(feedbackRepo.scoresBySessionID) != 0 {
		t.Fatalf("saved scores length = %d, want 0", len(feedbackRepo.scoresBySessionID))
	}
}

func TestSessionServicePassesScenarioHistoryAndUserInputToConversationAgent(t *testing.T) {
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
	sessionRepo := newFakeSessionRepository()
	created, err := sessionRepo.Create(model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusRunning,
		TurnCount:  1,
		CreatedAt:  time.Now(),
		Messages: []model.Message{
			{Role: model.MessageRoleUser, Content: "I study computer science.", Stage: "自我介绍"},
			{Role: model.MessageRoleAI, Content: "Could you tell me about a project?", Stage: "项目经历"},
		},
	})
	if err != nil {
		t.Fatalf("setup session returned error: %v", err)
	}
	conversation := &fakeConversationAgent{
		output: agent.ConversationOutput{
			Reply:    "What technical challenge did you solve?",
			Stage:    "技术追问",
			NextGoal: "ask user to explain a technical challenge",
		},
	}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(conversation),
	)

	result, err := sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   " I built a robot control project. ",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}

	if conversation.callCount != 1 {
		t.Fatalf("conversation call count = %d, want 1", conversation.callCount)
	}
	if conversation.input.Scenario.Code != "interview" {
		t.Fatalf("conversation scenario code = %q, want interview", conversation.input.Scenario.Code)
	}
	if conversation.input.UserContent != "I built a robot control project." {
		t.Fatalf("conversation user content = %q, want trimmed content", conversation.input.UserContent)
	}
	if len(conversation.input.History) != 2 {
		t.Fatalf("conversation history length = %d, want 2", len(conversation.input.History))
	}
	if result.AIMessage.Content != "What technical challenge did you solve?" {
		t.Fatalf("ai content = %q, want fake agent reply", result.AIMessage.Content)
	}
	if result.Stage != "技术追问" {
		t.Fatalf("stage = %q, want 技术追问", result.Stage)
	}
	if result.NextGoal != "ask user to explain a technical challenge" {
		t.Fatalf("next_goal = %q, want fake agent next goal", result.NextGoal)
	}
}

func TestSessionServiceReturnsConversationAgentFailureWithoutAppendingMessages(t *testing.T) {
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {ID: 1, Code: "interview"},
		},
	}
	sessionRepo := newFakeSessionRepository()
	created, err := sessionRepo.Create(model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusRunning,
		CreatedAt:  time.Now(),
		Messages:   []model.Message{},
	})
	if err != nil {
		t.Fatalf("setup session returned error: %v", err)
	}
	sessionService := service.NewSessionService(
		scenarioReader,
		sessionRepo,
		service.WithConversationAgent(&fakeConversationAgent{err: errors.New("llm failed")}),
	)

	_, err = sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "Hello",
	})

	if !errors.Is(err, service.ErrConversationAgentFailed) {
		t.Fatalf("error = %v, want ErrConversationAgentFailed", err)
	}
	saved, findErr := sessionRepo.FindByID(created.ID)
	if findErr != nil {
		t.Fatalf("FindByID returned error: %v", findErr)
	}
	if len(saved.Messages) != 0 {
		t.Fatalf("saved messages length = %d, want 0", len(saved.Messages))
	}
}

func TestSessionServiceRejectsBlankMessageContent(t *testing.T) {
	sessionService := service.NewSessionService(fakeScenarioReader{}, newFakeSessionRepository())

	_, err := sessionService.SendMessage(service.SendMessageInput{SessionID: 1, Content: "   "})

	if !errors.Is(err, service.ErrMessageContentRequired) {
		t.Fatalf("error = %v, want ErrMessageContentRequired", err)
	}
}

func TestSessionServiceRejectsMessageForFinishedSession(t *testing.T) {
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
	feedbackRepo := newFakeFeedbackRepository()
	correctionAgent := &fakeCorrectionAgent{}
	scoringAgent := &fakeScoringAgent{}
	sessionService := service.NewSessionService(
		fakeScenarioReader{},
		sessionRepo,
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(correctionAgent),
		service.WithScoringAgent(scoringAgent),
	)

	_, err = sessionService.SendMessage(service.SendMessageInput{
		SessionID: created.ID,
		Content:   "Hello",
	})

	if !errors.Is(err, service.ErrSessionAlreadyFinished) {
		t.Fatalf("error = %v, want ErrSessionAlreadyFinished", err)
	}
	if correctionAgent.callCount != 0 {
		t.Fatalf("correction call count = %d, want 0", correctionAgent.callCount)
	}
	if scoringAgent.callCount != 0 {
		t.Fatalf("scoring call count = %d, want 0", scoringAgent.callCount)
	}
	if feedbackRepo.callCount != 0 {
		t.Fatalf("feedback repo call count = %d, want 0", feedbackRepo.callCount)
	}
}

func setupFeedbackSession(t *testing.T) (fakeScenarioReader, *fakeSessionRepository, model.Session) {
	t.Helper()

	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:   1,
				Code: "interview",
				Stages: []model.ScenarioStage{
					{Name: "自我介绍"},
					{Name: "项目经历"},
				},
			},
		},
	}
	sessionRepo := newFakeSessionRepository()
	created, err := sessionRepo.Create(model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     model.SessionStatusRunning,
		CreatedAt:  time.Now(),
		Messages:   []model.Message{},
	})
	if err != nil {
		t.Fatalf("setup session returned error: %v", err)
	}

	return scenarioReader, sessionRepo, created
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
	nextID        int
	nextMessageID int
	createCount   int
	sessions      map[int]model.Session
}

func newFakeSessionRepository() *fakeSessionRepository {
	return &fakeSessionRepository{
		nextID:        1,
		nextMessageID: 1,
		sessions:      make(map[int]model.Session),
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

func (r *fakeSessionRepository) AppendTurn(id int, userMessage model.Message, aiMessage model.Message) (model.Session, error) {
	session, ok := r.sessions[id]
	if !ok {
		return model.Session{}, repository.ErrSessionNotFound
	}
	if session.Status == model.SessionStatusFinished {
		return model.Session{}, repository.ErrSessionAlreadyFinished
	}

	userMessage.ID = r.nextMessageID
	r.nextMessageID++
	userMessage.SessionID = id
	aiMessage.ID = r.nextMessageID
	r.nextMessageID++
	aiMessage.SessionID = id
	session.Messages = append(session.Messages, userMessage, aiMessage)
	session.TurnCount++
	r.sessions[id] = session

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

type fakeConversationAgent struct {
	output    agent.ConversationOutput
	err       error
	callCount int
	input     agent.ConversationInput
}

func (a *fakeConversationAgent) GenerateReply(ctx context.Context, input agent.ConversationInput) (agent.ConversationOutput, error) {
	a.callCount++
	a.input = input
	if a.err != nil {
		return agent.ConversationOutput{}, a.err
	}

	return a.output, nil
}

type fakeStreamingConversationAgent struct {
	output    agent.ConversationOutput
	err       error
	deltas    []string
	callCount int
	input     agent.ConversationInput
}

func (a *fakeStreamingConversationAgent) GenerateReply(ctx context.Context, input agent.ConversationInput) (agent.ConversationOutput, error) {
	a.callCount++
	a.input = input
	if a.err != nil {
		return agent.ConversationOutput{}, a.err
	}

	return a.output, nil
}

func (a *fakeStreamingConversationAgent) StreamReply(ctx context.Context, input agent.ConversationInput, onDelta func(agent.ConversationDelta) error) (agent.ConversationOutput, error) {
	a.callCount++
	a.input = input
	if a.err != nil {
		return agent.ConversationOutput{}, a.err
	}
	for _, delta := range a.deltas {
		if err := onDelta(agent.ConversationDelta{Content: delta}); err != nil {
			return agent.ConversationOutput{}, err
		}
	}

	return a.output, nil
}

type fakeCorrectionAgent struct {
	err       error
	callCount int
	input     agent.CorrectionInput
}

func (a *fakeCorrectionAgent) Correct(input agent.CorrectionInput) (agent.CorrectionOutput, error) {
	a.callCount++
	a.input = input
	if a.err != nil {
		return agent.CorrectionOutput{}, a.err
	}

	return agent.CorrectionOutput{
		Result: model.CorrectionResult{
			MessageID:     input.UserMessage.ID,
			SessionID:     input.UserMessage.SessionID,
			OriginalText:  input.UserMessage.Content,
			CorrectedText: "I am studying computer science, and I have done a project.",
			Errors: []model.CorrectionError{
				{
					Type:        model.CorrectionErrorTypeGrammar,
					Span:        "am study",
					Suggestion:  "am studying",
					Explanation: "be 动词后应接现在分词。",
				},
				{
					Type:        model.CorrectionErrorTypeGrammar,
					Span:        "have did",
					Suggestion:  "have done",
					Explanation: "现在完成时中 have 后应接过去分词 done。",
				},
			},
			BetterExpressions: []string{"I major in computer science."},
		},
	}, nil
}

type fakeScoringAgent struct {
	err       error
	callCount int
	input     agent.ScoringInput
}

func (a *fakeScoringAgent) Score(input agent.ScoringInput) (agent.ScoringOutput, error) {
	a.callCount++
	a.input = input
	if a.err != nil {
		return agent.ScoringOutput{}, a.err
	}

	return agent.ScoringOutput{
		Result: model.ScoreResult{
			MessageID:  input.Correction.MessageID,
			SessionID:  input.Correction.SessionID,
			Fluency:    75,
			Grammar:    72,
			Expression: 80,
			Vocabulary: 76,
			Completion: 85,
			TotalScore: 77,
			Comment:    "stable score",
		},
	}, nil
}

type fakeEventPublisher struct {
	events []stream.Event
	err    error
}

func (p *fakeEventPublisher) Publish(event stream.Event) error {
	p.events = append(p.events, event)

	return p.err
}

type fakeStateStore struct {
	err         error
	messages    map[int][]model.Message
	states      map[int]state.SessionState
	scores      map[int]model.ScoreResult
	corrections map[int][]model.CorrectionResult
	connections map[int]state.WebSocketConnectionState
}

func newFakeStateStore() *fakeStateStore {
	return &fakeStateStore{
		messages:    make(map[int][]model.Message),
		states:      make(map[int]state.SessionState),
		scores:      make(map[int]model.ScoreResult),
		corrections: make(map[int][]model.CorrectionResult),
		connections: make(map[int]state.WebSocketConnectionState),
	}
}

func (s *fakeStateStore) SaveMessageSnapshot(ctx context.Context, sessionID int, messages []model.Message) error {
	if s.err != nil {
		return s.err
	}
	s.messages[sessionID] = append([]model.Message(nil), messages...)
	return nil
}

func (s *fakeStateStore) GetMessageSnapshot(ctx context.Context, sessionID int) ([]model.Message, error) {
	messages, ok := s.messages[sessionID]
	if !ok {
		return nil, state.ErrStateNotFound
	}
	return append([]model.Message(nil), messages...), nil
}

func (s *fakeStateStore) SaveSessionState(ctx context.Context, sessionState state.SessionState) error {
	if s.err != nil {
		return s.err
	}
	s.states[sessionState.SessionID] = sessionState
	return nil
}

func (s *fakeStateStore) GetSessionState(ctx context.Context, sessionID int) (state.SessionState, error) {
	sessionState, ok := s.states[sessionID]
	if !ok {
		return state.SessionState{}, state.ErrStateNotFound
	}
	return sessionState, nil
}

func (s *fakeStateStore) SavePartialScore(ctx context.Context, score model.ScoreResult) error {
	if s.err != nil {
		return s.err
	}
	s.scores[score.SessionID] = score
	return nil
}

func (s *fakeStateStore) GetPartialScore(ctx context.Context, sessionID int) (model.ScoreResult, error) {
	score, ok := s.scores[sessionID]
	if !ok {
		return model.ScoreResult{}, state.ErrStateNotFound
	}
	return score, nil
}

func (s *fakeStateStore) AppendCorrection(ctx context.Context, correction model.CorrectionResult) error {
	if s.err != nil {
		return s.err
	}
	s.corrections[correction.SessionID] = append(s.corrections[correction.SessionID], correction)
	return nil
}

func (s *fakeStateStore) ListCorrections(ctx context.Context, sessionID int) ([]model.CorrectionResult, error) {
	corrections, ok := s.corrections[sessionID]
	if !ok {
		return nil, state.ErrStateNotFound
	}
	return append([]model.CorrectionResult(nil), corrections...), nil
}

func (s *fakeStateStore) SaveWebSocketConnection(ctx context.Context, connection state.WebSocketConnectionState) error {
	if s.err != nil {
		return s.err
	}
	s.connections[connection.SessionID] = connection
	return nil
}

func (s *fakeStateStore) GetWebSocketConnection(ctx context.Context, sessionID int) (state.WebSocketConnectionState, error) {
	connection, ok := s.connections[sessionID]
	if !ok {
		return state.WebSocketConnectionState{}, state.ErrStateNotFound
	}
	return connection, nil
}
