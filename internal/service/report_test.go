package service_test

import (
	"errors"
	"testing"
	"time"

	"speakmate/internal/agent"
	"speakmate/internal/model"
	"speakmate/internal/repository"
	"speakmate/internal/service"
	"speakmate/internal/stream"
)

func TestReportServiceGeneratesFinishedSessionReport(t *testing.T) {
	scenarioReader, sessionRepo, created := setupReportSession(t, model.SessionStatusFinished)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.correctionsBySessionID[created.ID] = []model.CorrectionResult{sampleCorrection(created.ID)}
	feedbackRepo.scoresBySessionID[created.ID] = sampleScore(created.ID)
	reportRepo := newFakeReportRepository()
	summaryAgent := &fakeSummaryAgent{
		output: agent.SummaryOutput{
			Summary:           "本次训练能够说明项目背景，但需要加强动词形式。",
			MajorProblems:     []string{"动词形式不稳定"},
			FrequentErrors:    []string{"am study -> am studying"},
			BetterExpressions: []string{"I major in computer science."},
			NextPracticePlan:  []string{"用 STAR 结构重写项目经历回答。"},
		},
	}
	reportService := service.NewReportService(
		scenarioReader,
		sessionRepo,
		feedbackRepo,
		reportRepo,
		service.WithSummaryAgent(summaryAgent),
		service.WithReportNow(func() time.Time {
			return time.Date(2026, 6, 7, 3, 5, 0, 0, time.UTC)
		}),
	)

	report, err := reportService.GenerateReport(created.ID)
	if err != nil {
		t.Fatalf("GenerateReport returned error: %v", err)
	}

	if summaryAgent.callCount != 1 {
		t.Fatalf("summary call count = %d, want 1", summaryAgent.callCount)
	}
	if summaryAgent.input.Scenario.Code != "interview" {
		t.Fatalf("summary scenario code = %q, want interview", summaryAgent.input.Scenario.Code)
	}
	if summaryAgent.input.Session.ID != created.ID {
		t.Fatalf("summary session id = %d, want %d", summaryAgent.input.Session.ID, created.ID)
	}
	if len(summaryAgent.input.Messages) != 2 {
		t.Fatalf("summary messages length = %d, want 2", len(summaryAgent.input.Messages))
	}
	if len(summaryAgent.input.Corrections) != 1 {
		t.Fatalf("summary corrections length = %d, want 1", len(summaryAgent.input.Corrections))
	}
	if summaryAgent.input.Score.TotalScore != 77 {
		t.Fatalf("summary score total = %d, want 77", summaryAgent.input.Score.TotalScore)
	}

	if report.SessionID != created.ID {
		t.Fatalf("report session id = %d, want %d", report.SessionID, created.ID)
	}
	if report.Scenario.Code != "interview" {
		t.Fatalf("report scenario code = %q, want interview", report.Scenario.Code)
	}
	if report.DurationSeconds != 180 {
		t.Fatalf("duration_seconds = %d, want 180", report.DurationSeconds)
	}
	if report.TurnCount != 1 {
		t.Fatalf("turn_count = %d, want 1", report.TurnCount)
	}
	if report.TotalScore != 77 {
		t.Fatalf("total_score = %d, want 77", report.TotalScore)
	}
	if report.Summary != "本次训练能够说明项目背景，但需要加强动词形式。" {
		t.Fatalf("summary = %q, want summary agent output", report.Summary)
	}
	if len(report.FrequentErrors) != 1 || report.FrequentErrors[0] != "am study -> am studying" {
		t.Fatalf("frequent_errors = %#v, want summary agent output", report.FrequentErrors)
	}
	if report.CreatedAt.Format(time.RFC3339) != "2026-06-07T03:05:00Z" {
		t.Fatalf("created_at = %s, want fixed time", report.CreatedAt.Format(time.RFC3339))
	}

	saved, err := reportRepo.FindBySessionID(created.ID)
	if err != nil {
		t.Fatalf("FindBySessionID returned error: %v", err)
	}
	if saved.Summary != report.Summary {
		t.Fatalf("saved summary = %q, want generated report summary", saved.Summary)
	}
}

func TestReportServicePublishesReportDoneEvent(t *testing.T) {
	scenarioReader, sessionRepo, created := setupReportSession(t, model.SessionStatusFinished)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.correctionsBySessionID[created.ID] = []model.CorrectionResult{sampleCorrection(created.ID)}
	feedbackRepo.scoresBySessionID[created.ID] = sampleScore(created.ID)
	publisher := &fakeEventPublisher{}
	reportService := service.NewReportService(
		scenarioReader,
		sessionRepo,
		feedbackRepo,
		newFakeReportRepository(),
		service.WithSummaryAgent(&fakeSummaryAgent{
			output: agent.SummaryOutput{
				Summary:          "本次训练能够说明项目背景。",
				MajorProblems:    []string{"动词形式不稳定"},
				FrequentErrors:   []string{"am study -> am studying"},
				NextPracticePlan: []string{"复述项目经历。"},
			},
		}),
		service.WithReportEventPublisher(publisher),
	)

	report, err := reportService.GenerateReport(created.ID)
	if err != nil {
		t.Fatalf("GenerateReport returned error: %v", err)
	}

	if len(publisher.events) != 1 {
		t.Fatalf("published events length = %d, want 1: %+v", len(publisher.events), publisher.events)
	}
	event := publisher.events[0]
	if event.Type != stream.EventTypeReportDone {
		t.Fatalf("event type = %q, want %q", event.Type, stream.EventTypeReportDone)
	}
	if event.SessionID != created.ID {
		t.Fatalf("event session id = %d, want %d", event.SessionID, created.ID)
	}
	payload, ok := event.Payload.(stream.ReportDonePayload)
	if !ok {
		t.Fatalf("payload type = %T, want ReportDonePayload", event.Payload)
	}
	if payload.TotalScore != report.TotalScore || payload.Summary != report.Summary {
		t.Fatalf("payload = %+v, want report total/summary", payload)
	}
}

func TestReportServicePublishesErrorEventWhenGenerationFails(t *testing.T) {
	scenarioReader, sessionRepo, created := setupReportSession(t, model.SessionStatusFinished)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.correctionsBySessionID[created.ID] = []model.CorrectionResult{sampleCorrection(created.ID)}
	feedbackRepo.scoresBySessionID[created.ID] = sampleScore(created.ID)
	publisher := &fakeEventPublisher{}
	reportService := service.NewReportService(
		scenarioReader,
		sessionRepo,
		feedbackRepo,
		newFakeReportRepository(),
		service.WithSummaryAgent(&fakeSummaryAgent{err: errors.New("summary failed")}),
		service.WithReportEventPublisher(publisher),
	)

	_, err := reportService.GenerateReport(created.ID)

	if !errors.Is(err, service.ErrSummaryAgentFailed) {
		t.Fatalf("GenerateReport error = %v, want ErrSummaryAgentFailed", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events length = %d, want 1: %+v", len(publisher.events), publisher.events)
	}
	event := publisher.events[0]
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
	if payload.Code != "summary_agent_failed" || payload.Message != "summary agent failed" {
		t.Fatalf("payload = %+v, want summary failure", payload)
	}
}

func TestReportServiceGetReportReturnsSavedReport(t *testing.T) {
	reportRepo := newFakeReportRepository()
	saved := model.Report{SessionID: 7, Summary: "saved report"}
	if err := reportRepo.Save(saved); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	reportService := service.NewReportService(fakeScenarioReader{}, newFakeSessionRepository(), newFakeFeedbackRepository(), reportRepo)

	found, err := reportService.GetReport(7)
	if err != nil {
		t.Fatalf("GetReport returned error: %v", err)
	}

	if found.Summary != "saved report" {
		t.Fatalf("summary = %q, want saved report", found.Summary)
	}
}

func TestReportServiceGenerateReportOverwritesExistingReport(t *testing.T) {
	scenarioReader, sessionRepo, created := setupReportSession(t, model.SessionStatusFinished)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.correctionsBySessionID[created.ID] = []model.CorrectionResult{sampleCorrection(created.ID)}
	feedbackRepo.scoresBySessionID[created.ID] = sampleScore(created.ID)
	reportRepo := newFakeReportRepository()
	summaryAgent := &fakeSummaryAgent{
		outputs: []agent.SummaryOutput{
			{Summary: "first", MajorProblems: []string{"first"}, FrequentErrors: []string{"first"}, BetterExpressions: []string{"first"}, NextPracticePlan: []string{"first"}},
			{Summary: "second", MajorProblems: []string{"second"}, FrequentErrors: []string{"second"}, BetterExpressions: []string{"second"}, NextPracticePlan: []string{"second"}},
		},
	}
	reportService := service.NewReportService(
		scenarioReader,
		sessionRepo,
		feedbackRepo,
		reportRepo,
		service.WithSummaryAgent(summaryAgent),
	)

	if _, err := reportService.GenerateReport(created.ID); err != nil {
		t.Fatalf("first GenerateReport returned error: %v", err)
	}
	second, err := reportService.GenerateReport(created.ID)
	if err != nil {
		t.Fatalf("second GenerateReport returned error: %v", err)
	}

	if second.Summary != "second" {
		t.Fatalf("second summary = %q, want overwritten summary", second.Summary)
	}
	found, err := reportService.GetReport(created.ID)
	if err != nil {
		t.Fatalf("GetReport returned error: %v", err)
	}
	if found.Summary != "second" {
		t.Fatalf("saved summary = %q, want overwritten summary", found.Summary)
	}
}

func TestReportServiceRejectsInvalidSessionID(t *testing.T) {
	reportRepo := newFakeReportRepository()
	reportService := service.NewReportService(fakeScenarioReader{}, newFakeSessionRepository(), newFakeFeedbackRepository(), reportRepo)

	if _, err := reportService.GenerateReport(0); !errors.Is(err, service.ErrInvalidReportRequest) {
		t.Fatalf("GenerateReport error = %v, want ErrInvalidReportRequest", err)
	}
	if _, err := reportService.GetReport(0); !errors.Is(err, service.ErrInvalidReportRequest) {
		t.Fatalf("GetReport error = %v, want ErrInvalidReportRequest", err)
	}
}

func TestReportServiceReturnsSessionNotFound(t *testing.T) {
	reportService := service.NewReportService(fakeScenarioReader{}, newFakeSessionRepository(), newFakeFeedbackRepository(), newFakeReportRepository())

	_, err := reportService.GenerateReport(404)

	if !errors.Is(err, service.ErrSessionNotFound) {
		t.Fatalf("GenerateReport error = %v, want ErrSessionNotFound", err)
	}
}

func TestReportServiceRequiresFinishedSession(t *testing.T) {
	scenarioReader, sessionRepo, created := setupReportSession(t, model.SessionStatusRunning)
	reportService := service.NewReportService(scenarioReader, sessionRepo, newFakeFeedbackRepository(), newFakeReportRepository())

	_, err := reportService.GenerateReport(created.ID)

	if !errors.Is(err, service.ErrSessionNotFinished) {
		t.Fatalf("GenerateReport error = %v, want ErrSessionNotFinished", err)
	}
}

func TestReportServiceRequiresCorrectionsAndScore(t *testing.T) {
	scenarioReader, sessionRepo, created := setupReportSession(t, model.SessionStatusFinished)

	t.Run("missing corrections", func(t *testing.T) {
		feedbackRepo := newFakeFeedbackRepository()
		feedbackRepo.scoresBySessionID[created.ID] = sampleScore(created.ID)
		reportService := service.NewReportService(scenarioReader, sessionRepo, feedbackRepo, newFakeReportRepository())

		_, err := reportService.GenerateReport(created.ID)

		if !errors.Is(err, service.ErrReportFeedbackMissing) {
			t.Fatalf("GenerateReport error = %v, want ErrReportFeedbackMissing", err)
		}
	})

	t.Run("missing score", func(t *testing.T) {
		feedbackRepo := newFakeFeedbackRepository()
		feedbackRepo.correctionsBySessionID[created.ID] = []model.CorrectionResult{sampleCorrection(created.ID)}
		reportService := service.NewReportService(scenarioReader, sessionRepo, feedbackRepo, newFakeReportRepository())

		_, err := reportService.GenerateReport(created.ID)

		if !errors.Is(err, service.ErrReportFeedbackMissing) {
			t.Fatalf("GenerateReport error = %v, want ErrReportFeedbackMissing", err)
		}
	})
}

func TestReportServiceReturnsSummaryAgentFailure(t *testing.T) {
	scenarioReader, sessionRepo, created := setupReportSession(t, model.SessionStatusFinished)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.correctionsBySessionID[created.ID] = []model.CorrectionResult{sampleCorrection(created.ID)}
	feedbackRepo.scoresBySessionID[created.ID] = sampleScore(created.ID)
	reportService := service.NewReportService(
		scenarioReader,
		sessionRepo,
		feedbackRepo,
		newFakeReportRepository(),
		service.WithSummaryAgent(&fakeSummaryAgent{err: errors.New("summary failed")}),
	)

	_, err := reportService.GenerateReport(created.ID)

	if !errors.Is(err, service.ErrSummaryAgentFailed) {
		t.Fatalf("GenerateReport error = %v, want ErrSummaryAgentFailed", err)
	}
}

func setupReportSession(t *testing.T, status model.SessionStatus) (fakeScenarioReader, *fakeSessionRepository, model.Session) {
	t.Helper()

	createdAt := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)
	endedAt := createdAt.Add(3 * time.Minute)
	scenarioReader := fakeScenarioReader{
		scenarios: map[int]model.Scenario{
			1: {
				ID:         1,
				Code:       "interview",
				Name:       "英语面试",
				Difficulty: "medium",
				UserGoal:   "清晰介绍项目经历",
			},
		},
	}
	sessionRepo := newFakeSessionRepository()
	session := model.Session{
		ScenarioID: 1,
		UserID:     1,
		Status:     status,
		TurnCount:  1,
		CreatedAt:  createdAt,
		Messages: []model.Message{
			{ID: 1, Role: model.MessageRoleUser, Content: "I am study computer science.", Stage: "自我介绍", CreatedAt: createdAt},
			{ID: 2, Role: model.MessageRoleAI, Content: "What was your project?", Stage: "项目经历", CreatedAt: createdAt},
		},
	}
	if status == model.SessionStatusFinished {
		session.EndedAt = &endedAt
	}
	created, err := sessionRepo.Create(session)
	if err != nil {
		t.Fatalf("setup session returned error: %v", err)
	}
	for i := range created.Messages {
		created.Messages[i].SessionID = created.ID
	}
	sessionRepo.sessions[created.ID] = created

	return scenarioReader, sessionRepo, created
}

func sampleCorrection(sessionID int) model.CorrectionResult {
	return model.CorrectionResult{
		MessageID:     1,
		SessionID:     sessionID,
		OriginalText:  "I am study computer science.",
		CorrectedText: "I am studying computer science.",
		Errors: []model.CorrectionError{
			{
				Type:        model.CorrectionErrorTypeGrammar,
				Span:        "am study",
				Suggestion:  "am studying",
				Explanation: "be 动词后应接现在分词。",
			},
		},
		BetterExpressions: []string{"I major in computer science."},
	}
}

func sampleScore(sessionID int) model.ScoreResult {
	return model.ScoreResult{
		MessageID:  1,
		SessionID:  sessionID,
		Fluency:    75,
		Grammar:    72,
		Expression: 80,
		Vocabulary: 76,
		Completion: 85,
		TotalScore: 77,
		Comment:    "用户能够表达核心意思，但存在时态和动词形式错误。",
	}
}

type fakeReportRepository struct {
	reports map[int]model.Report
	err     error
}

func newFakeReportRepository() *fakeReportRepository {
	return &fakeReportRepository{
		reports: make(map[int]model.Report),
	}
}

func (r *fakeReportRepository) Save(report model.Report) error {
	if r.err != nil {
		return r.err
	}
	r.reports[report.SessionID] = report

	return nil
}

func (r *fakeReportRepository) FindBySessionID(sessionID int) (model.Report, error) {
	if r.err != nil {
		return model.Report{}, r.err
	}
	report, ok := r.reports[sessionID]
	if !ok {
		return model.Report{}, repository.ErrReportNotFound
	}

	return report, nil
}

type fakeSummaryAgent struct {
	output    agent.SummaryOutput
	outputs   []agent.SummaryOutput
	err       error
	callCount int
	input     agent.SummaryInput
}

func (a *fakeSummaryAgent) Summarize(input agent.SummaryInput) (agent.SummaryOutput, error) {
	a.callCount++
	a.input = input
	if a.err != nil {
		return agent.SummaryOutput{}, a.err
	}
	if len(a.outputs) >= a.callCount {
		return a.outputs[a.callCount-1], nil
	}

	return a.output, nil
}
