package service_test

import (
	"errors"
	"testing"
	"time"

	"speakmate/internal/model"
	"speakmate/internal/service"
)

func TestHistoryInsightsEmptyHistoryReturnsEmptyResult(t *testing.T) {
	now := fixedHistoryInsightsNow()
	sessionRepo := newFakeHistoryInsightSessionRepository()
	insightsService := newHistoryInsightsTestService(t, sessionRepo, nil, nil, nil, now)

	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	if result.Summary.Days != 30 {
		t.Fatalf("summary days = %d, want default 30", result.Summary.Days)
	}
	if result.Summary.TotalSessions != 0 || result.Summary.FinishedSessions != 0 ||
		result.Summary.RunningSessions != 0 || result.Summary.ScoredSessions != 0 ||
		result.Summary.GeneratedReports != 0 {
		t.Fatalf("summary counts = %+v, want all zero", result.Summary)
	}
	if result.Summary.AverageScore != nil || result.Summary.PreviousAverageScore != nil || result.Summary.ScoreDelta != nil {
		t.Fatalf("summary scores = %+v, want nil score pointers", result.Summary)
	}
	if len(result.ScoreTrend) != 0 {
		t.Fatalf("score trend length = %d, want 0", len(result.ScoreTrend))
	}
	if len(result.ScenarioTrends) != 0 {
		t.Fatalf("scenario trends length = %d, want 0", len(result.ScenarioTrends))
	}
	if len(result.FrequentErrors) != 0 {
		t.Fatalf("frequent errors length = %d, want 0", len(result.FrequentErrors))
	}
	if result.NextRecommendation != nil {
		t.Fatalf("next recommendation = %+v, want nil", result.NextRecommendation)
	}
}

func TestHistoryInsightsRunningSessionsCountTowardTotalButNotAverages(t *testing.T) {
	now := fixedHistoryInsightsNow()
	currentStart := now.AddDate(0, 0, -30)
	previousStart := now.AddDate(0, 0, -60)
	finished := historyInsightSession(1, 1, model.SessionStatusFinished, now.AddDate(0, 0, -2))
	running := historyInsightSession(2, 1, model.SessionStatusRunning, now.AddDate(0, 0, -1))
	sessionRepo := newFakeHistoryInsightSessionRepository()
	sessionRepo.setWindow(currentStart, now, []model.Session{running, finished})
	sessionRepo.setWindow(previousStart, currentStart, nil)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.scoresBySessionID[finished.ID] = historyInsightScore(finished.ID, 90)
	feedbackRepo.scoresBySessionID[running.ID] = historyInsightScore(running.ID, 20)
	insightsService := newHistoryInsightsTestService(t, sessionRepo, feedbackRepo, nil, nil, now)

	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1, Days: 30})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	if result.Summary.TotalSessions != 2 {
		t.Fatalf("total sessions = %d, want 2", result.Summary.TotalSessions)
	}
	if result.Summary.FinishedSessions != 1 {
		t.Fatalf("finished sessions = %d, want 1", result.Summary.FinishedSessions)
	}
	if result.Summary.RunningSessions != 1 {
		t.Fatalf("running sessions = %d, want 1", result.Summary.RunningSessions)
	}
	if result.Summary.ScoredSessions != 1 {
		t.Fatalf("scored sessions = %d, want 1", result.Summary.ScoredSessions)
	}
	assertHistoryInsightIntPtr(t, result.Summary.AverageScore, 90, "average score")
	if len(result.ScoreTrend) != 1 {
		t.Fatalf("score trend length = %d, want 1: %+v", len(result.ScoreTrend), result.ScoreTrend)
	}
	if result.ScoreTrend[0].AverageScore != 90 || result.ScoreTrend[0].SessionCount != 1 {
		t.Fatalf("score trend point = %+v, want average 90 from one scored finished session", result.ScoreTrend[0])
	}
}

func TestHistoryInsightsCalculatesCurrentPreviousAveragesAndDelta(t *testing.T) {
	now := fixedHistoryInsightsNow()
	currentStart := now.AddDate(0, 0, -30)
	previousStart := now.AddDate(0, 0, -60)
	currentA := historyInsightSession(1, 1, model.SessionStatusFinished, now.AddDate(0, 0, -1))
	currentB := historyInsightSession(2, 1, model.SessionStatusFinished, now.AddDate(0, 0, -2))
	previousA := historyInsightSession(3, 1, model.SessionStatusFinished, now.AddDate(0, 0, -31))
	previousB := historyInsightSession(4, 1, model.SessionStatusFinished, now.AddDate(0, 0, -32))
	sessionRepo := newFakeHistoryInsightSessionRepository()
	sessionRepo.setWindow(currentStart, now, []model.Session{currentA, currentB})
	sessionRepo.setWindow(previousStart, currentStart, []model.Session{previousA, previousB})
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.scoresBySessionID[currentA.ID] = historyInsightScore(currentA.ID, 80)
	feedbackRepo.scoresBySessionID[currentB.ID] = historyInsightScore(currentB.ID, 81)
	feedbackRepo.scoresBySessionID[previousA.ID] = historyInsightScore(previousA.ID, 70)
	feedbackRepo.scoresBySessionID[previousB.ID] = historyInsightScore(previousB.ID, 70)
	insightsService := newHistoryInsightsTestService(t, sessionRepo, feedbackRepo, nil, nil, now)

	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1, Days: 30})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	assertHistoryInsightIntPtr(t, result.Summary.AverageScore, 81, "current average score")
	assertHistoryInsightIntPtr(t, result.Summary.PreviousAverageScore, 70, "previous average score")
	assertHistoryInsightIntPtr(t, result.Summary.ScoreDelta, 11, "score delta")
}

func TestHistoryInsightsUsesSevenAndThirtyDayWindows(t *testing.T) {
	now := fixedHistoryInsightsNow()
	tests := []struct {
		name          string
		days          int
		currentStart  time.Time
		previousStart time.Time
	}{
		{
			name:          "seven days",
			days:          7,
			currentStart:  now.AddDate(0, 0, -7),
			previousStart: now.AddDate(0, 0, -14),
		},
		{
			name:          "thirty days",
			days:          30,
			currentStart:  now.AddDate(0, 0, -30),
			previousStart: now.AddDate(0, 0, -60),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := newFakeHistoryInsightSessionRepository()
			insightsService := newHistoryInsightsTestService(t, sessionRepo, nil, nil, nil, now)

			_, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 42, Days: tt.days})
			if err != nil {
				t.Fatalf("GetInsights returned error: %v", err)
			}

			if len(sessionRepo.queries) != 2 {
				t.Fatalf("query count = %d, want 2: %+v", len(sessionRepo.queries), sessionRepo.queries)
			}
			assertHistoryInsightWindow(t, sessionRepo.queries[0], 42, tt.currentStart, now)
			assertHistoryInsightWindow(t, sessionRepo.queries[1], 42, tt.previousStart, tt.currentStart)
		})
	}
}

func TestHistoryInsightsBuildsScenarioTrendFirstLatestAndDelta(t *testing.T) {
	now := fixedHistoryInsightsNow()
	currentStart := now.AddDate(0, 0, -30)
	previousStart := now.AddDate(0, 0, -60)
	older := historyInsightSession(1, 1, model.SessionStatusFinished, now.AddDate(0, 0, -8))
	newer := historyInsightSession(2, 1, model.SessionStatusFinished, now.AddDate(0, 0, -2))
	other := historyInsightSession(3, 2, model.SessionStatusFinished, now.AddDate(0, 0, -1))
	sessionRepo := newFakeHistoryInsightSessionRepository()
	sessionRepo.setWindow(currentStart, now, []model.Session{other, newer, older})
	sessionRepo.setWindow(previousStart, currentStart, nil)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.scoresBySessionID[older.ID] = historyInsightScore(older.ID, 60)
	feedbackRepo.scoresBySessionID[newer.ID] = historyInsightScore(newer.ID, 85)
	feedbackRepo.scoresBySessionID[other.ID] = historyInsightScore(other.ID, 70)
	insightsService := newHistoryInsightsTestService(t, sessionRepo, feedbackRepo, nil, nil, now)

	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1, Days: 30})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	trend := findHistoryScenarioTrend(t, result.ScenarioTrends, 1)
	if trend.SessionCount != 2 {
		t.Fatalf("scenario session count = %d, want 2", trend.SessionCount)
	}
	if trend.ScoredSessions != 2 {
		t.Fatalf("scenario scored sessions = %d, want 2", trend.ScoredSessions)
	}
	assertHistoryInsightIntPtr(t, trend.AverageScore, 73, "scenario average score")
	assertHistoryInsightIntPtr(t, trend.FirstScore, 60, "scenario first score")
	assertHistoryInsightIntPtr(t, trend.LatestScore, 85, "scenario latest score")
	assertHistoryInsightIntPtr(t, trend.ScoreDelta, 25, "scenario score delta")
	if !trend.LastTrainedAt.Equal(newer.CreatedAt) {
		t.Fatalf("last trained at = %s, want %s", trend.LastTrainedAt, newer.CreatedAt)
	}
}

func TestHistoryInsightsScenarioTrendsOnlyIncludeScoredScenarios(t *testing.T) {
	now := fixedHistoryInsightsNow()
	currentStart := now.AddDate(0, 0, -30)
	previousStart := now.AddDate(0, 0, -60)
	scored := historyInsightSession(1, 1, model.SessionStatusFinished, now.AddDate(0, 0, -3))
	runningOnly := historyInsightSession(2, 2, model.SessionStatusRunning, now.AddDate(0, 0, -2))
	unscoredFinished := historyInsightSession(3, 3, model.SessionStatusFinished, now.AddDate(0, 0, -1))
	sessionRepo := newFakeHistoryInsightSessionRepository()
	sessionRepo.setWindow(currentStart, now, []model.Session{unscoredFinished, runningOnly, scored})
	sessionRepo.setWindow(previousStart, currentStart, nil)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.scoresBySessionID[scored.ID] = historyInsightScore(scored.ID, 76)
	scenarios := map[int]model.Scenario{
		1: {ID: 1, Code: "interview", Name: "Interview", Difficulty: "medium"},
		2: {ID: 2, Code: "presentation", Name: "Presentation", Difficulty: "hard"},
		3: {ID: 3, Code: "small-talk", Name: "Small Talk", Difficulty: "easy"},
	}
	insightsService := newHistoryInsightsTestService(t, sessionRepo, feedbackRepo, nil, scenarios, now)

	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1, Days: 30})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	if result.Summary.TotalSessions != 3 {
		t.Fatalf("total sessions = %d, want 3", result.Summary.TotalSessions)
	}
	if result.Summary.RunningSessions != 1 {
		t.Fatalf("running sessions = %d, want 1", result.Summary.RunningSessions)
	}
	if result.Summary.FinishedSessions != 2 {
		t.Fatalf("finished sessions = %d, want 2", result.Summary.FinishedSessions)
	}
	if result.Summary.ScoredSessions != 1 {
		t.Fatalf("scored sessions = %d, want 1", result.Summary.ScoredSessions)
	}
	if len(result.ScenarioTrends) != 1 {
		t.Fatalf("scenario trends length = %d, want only scored scenario: %+v", len(result.ScenarioTrends), result.ScenarioTrends)
	}
	if result.ScenarioTrends[0].Scenario.ID != scored.ScenarioID {
		t.Fatalf("scenario trend id = %d, want scored scenario %d", result.ScenarioTrends[0].Scenario.ID, scored.ScenarioID)
	}
	assertHistoryInsightIntPtr(t, result.ScenarioTrends[0].AverageScore, 76, "scenario average score")
}

func TestHistoryInsightsScenarioTrendFirstLatestScoreUsesIDTieBreaks(t *testing.T) {
	now := fixedHistoryInsightsNow()
	currentStart := now.AddDate(0, 0, -30)
	previousStart := now.AddDate(0, 0, -60)
	createdAt := now.AddDate(0, 0, -4)
	lowerID := historyInsightSession(10, 1, model.SessionStatusFinished, createdAt)
	middleID := historyInsightSession(20, 1, model.SessionStatusFinished, createdAt)
	higherID := historyInsightSession(30, 1, model.SessionStatusFinished, createdAt)
	sessionRepo := newFakeHistoryInsightSessionRepository()
	sessionRepo.setWindow(currentStart, now, []model.Session{middleID, higherID, lowerID})
	sessionRepo.setWindow(previousStart, currentStart, nil)
	feedbackRepo := newFakeFeedbackRepository()
	feedbackRepo.scoresBySessionID[lowerID.ID] = historyInsightScore(lowerID.ID, 60)
	feedbackRepo.scoresBySessionID[middleID.ID] = historyInsightScore(middleID.ID, 70)
	feedbackRepo.scoresBySessionID[higherID.ID] = historyInsightScore(higherID.ID, 84)
	insightsService := newHistoryInsightsTestService(t, sessionRepo, feedbackRepo, nil, nil, now)

	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1, Days: 30})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	trend := findHistoryScenarioTrend(t, result.ScenarioTrends, 1)
	assertHistoryInsightIntPtr(t, trend.FirstScore, 60, "scenario first score")
	assertHistoryInsightIntPtr(t, trend.LatestScore, 84, "scenario latest score")
	assertHistoryInsightIntPtr(t, trend.ScoreDelta, 24, "scenario score delta")
}

func TestHistoryInsightsAggregatesFrequentErrors(t *testing.T) {
	now := fixedHistoryInsightsNow()
	currentStart := now.AddDate(0, 0, -30)
	previousStart := now.AddDate(0, 0, -60)
	sessions := []model.Session{
		historyInsightSession(1, 1, model.SessionStatusFinished, now.AddDate(0, 0, -8)),
		historyInsightSession(2, 1, model.SessionStatusFinished, now.AddDate(0, 0, -1)),
		historyInsightSession(3, 1, model.SessionStatusFinished, now.AddDate(0, 0, -2)),
		historyInsightSession(4, 1, model.SessionStatusFinished, now.AddDate(0, 0, -3)),
		historyInsightSession(5, 1, model.SessionStatusFinished, now.AddDate(0, 0, -4)),
		historyInsightSession(6, 1, model.SessionStatusFinished, now.AddDate(0, 0, -5)),
	}
	sessionRepo := newFakeHistoryInsightSessionRepository()
	sessionRepo.setWindow(currentStart, now, sessions)
	sessionRepo.setWindow(previousStart, currentStart, nil)
	reportRepo := newFakeReportRepository()
	reportRepo.reports[1] = model.Report{
		SessionID:      1,
		FrequentErrors: []string{" Am Study -> am studying ", "have did -> have done", "\u6682\u672a\u53d1\u73b0\u9ad8\u9891\u9519\u8bef"},
		CreatedAt:      sessions[0].CreatedAt,
	}
	reportRepo.reports[2] = model.Report{
		SessionID:      2,
		FrequentErrors: []string{"AM STUDY -> am studying", "article missing"},
		CreatedAt:      sessions[1].CreatedAt,
	}
	reportRepo.reports[3] = model.Report{
		SessionID:      3,
		FrequentErrors: []string{"word choice -> use precise verbs"},
		CreatedAt:      sessions[2].CreatedAt,
	}
	reportRepo.reports[4] = model.Report{
		SessionID:      4,
		FrequentErrors: []string{"preposition on -> preposition in"},
		CreatedAt:      sessions[3].CreatedAt,
	}
	reportRepo.reports[5] = model.Report{
		SessionID:      5,
		FrequentErrors: []string{"tense drift"},
		CreatedAt:      sessions[4].CreatedAt,
	}
	reportRepo.reports[6] = model.Report{
		SessionID:      6,
		FrequentErrors: []string{"filler words"},
		CreatedAt:      sessions[5].CreatedAt,
	}
	insightsService := newHistoryInsightsTestService(t, sessionRepo, nil, reportRepo, nil, now)

	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1, Days: 30})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	if result.Summary.GeneratedReports != 6 {
		t.Fatalf("generated reports = %d, want 6", result.Summary.GeneratedReports)
	}
	if len(result.FrequentErrors) != 5 {
		t.Fatalf("frequent error length = %d, want capped 5: %+v", len(result.FrequentErrors), result.FrequentErrors)
	}
	wantKeys := []string{"am study", "article missing", "word choice", "preposition on", "tense drift"}
	for i, wantKey := range wantKeys {
		if result.FrequentErrors[i].Key != wantKey {
			t.Fatalf("frequent error %d key = %q, want %q; errors = %+v", i, result.FrequentErrors[i].Key, wantKey, result.FrequentErrors)
		}
	}
	repeated := result.FrequentErrors[0]
	if repeated.Count != 2 {
		t.Fatalf("repeated error count = %d, want 2", repeated.Count)
	}
	if repeated.Category != "grammar" {
		t.Fatalf("repeated error category = %q, want grammar", repeated.Category)
	}
	if repeated.Suggestion != "am studying" {
		t.Fatalf("repeated error suggestion = %q, want am studying", repeated.Suggestion)
	}
	if repeated.LatestEvidence != "AM STUDY -> am studying" {
		t.Fatalf("latest evidence = %q, want latest source string", repeated.LatestEvidence)
	}
	if !repeated.LastSeenAt.Equal(sessions[1].CreatedAt) {
		t.Fatalf("last seen at = %s, want %s", repeated.LastSeenAt, sessions[1].CreatedAt)
	}
	if repeated.SourceSessionID != sessions[1].ID {
		t.Fatalf("source session id = %d, want %d", repeated.SourceSessionID, sessions[1].ID)
	}
	if result.FrequentErrors[1].Suggestion != "" {
		t.Fatalf("no-arrow suggestion = %q, want empty", result.FrequentErrors[1].Suggestion)
	}
}

func TestHistoryInsightsRecommendationPriority(t *testing.T) {
	t.Run("repeated error beats weakest scenario and running session", func(t *testing.T) {
		now := fixedHistoryInsightsNow()
		s1 := historyInsightSession(1, 1, model.SessionStatusFinished, now.AddDate(0, 0, -5))
		s2 := historyInsightSession(2, 1, model.SessionStatusFinished, now.AddDate(0, 0, -1))
		weak := historyInsightSession(3, 2, model.SessionStatusFinished, now.AddDate(0, 0, -2))
		running := historyInsightSession(4, 2, model.SessionStatusRunning, now.AddDate(0, 0, -1))
		feedbackRepo := newFakeFeedbackRepository()
		feedbackRepo.scoresBySessionID[s1.ID] = historyInsightScore(s1.ID, 95)
		feedbackRepo.scoresBySessionID[s2.ID] = historyInsightScore(s2.ID, 90)
		feedbackRepo.scoresBySessionID[weak.ID] = historyInsightScore(weak.ID, 35)
		reportRepo := newFakeReportRepository()
		reportRepo.reports[s1.ID] = model.Report{SessionID: s1.ID, FrequentErrors: []string{"am study -> am studying"}, CreatedAt: s1.CreatedAt}
		reportRepo.reports[s2.ID] = model.Report{SessionID: s2.ID, FrequentErrors: []string{"AM STUDY -> am studying"}, CreatedAt: s2.CreatedAt}
		result := runHistoryInsightsRecommendationCase(t, now, []model.Session{running, s2, weak, s1}, feedbackRepo, reportRepo)

		if result.NextRecommendation == nil {
			t.Fatal("next recommendation is nil, want repeated error recommendation")
		}
		if result.NextRecommendation.Type != "scenario_repractice" {
			t.Fatalf("recommendation type = %q, want scenario_repractice", result.NextRecommendation.Type)
		}
		if result.NextRecommendation.SessionID != s2.ID {
			t.Fatalf("recommendation session id = %d, want newest repeated-error session %d", result.NextRecommendation.SessionID, s2.ID)
		}
		if result.NextRecommendation.Scenario == nil || result.NextRecommendation.Scenario.ID != s2.ScenarioID {
			t.Fatalf("recommendation scenario = %+v, want scenario %d", result.NextRecommendation.Scenario, s2.ScenarioID)
		}
		if result.NextRecommendation.Focus != "am study" {
			t.Fatalf("recommendation focus = %q, want am study", result.NextRecommendation.Focus)
		}
		if result.NextRecommendation.Reason == "" {
			t.Fatal("recommendation reason is empty")
		}
	})

	t.Run("weakest scored scenario is recommended without repeated errors", func(t *testing.T) {
		now := fixedHistoryInsightsNow()
		strong := historyInsightSession(11, 1, model.SessionStatusFinished, now.AddDate(0, 0, -2))
		weak := historyInsightSession(12, 2, model.SessionStatusFinished, now.AddDate(0, 0, -1))
		feedbackRepo := newFakeFeedbackRepository()
		feedbackRepo.scoresBySessionID[strong.ID] = historyInsightScore(strong.ID, 88)
		feedbackRepo.scoresBySessionID[weak.ID] = historyInsightScore(weak.ID, 55)
		result := runHistoryInsightsRecommendationCase(t, now, []model.Session{weak, strong}, feedbackRepo, newFakeReportRepository())

		if result.NextRecommendation == nil {
			t.Fatal("next recommendation is nil, want weakest scenario recommendation")
		}
		if result.NextRecommendation.Type != "scenario_repractice" {
			t.Fatalf("recommendation type = %q, want scenario_repractice", result.NextRecommendation.Type)
		}
		if result.NextRecommendation.Scenario == nil || result.NextRecommendation.Scenario.ID != weak.ScenarioID {
			t.Fatalf("recommendation scenario = %+v, want scenario %d", result.NextRecommendation.Scenario, weak.ScenarioID)
		}
		if result.NextRecommendation.SessionID != weak.ID {
			t.Fatalf("recommendation session id = %d, want weak scenario latest session %d", result.NextRecommendation.SessionID, weak.ID)
		}
	})

	t.Run("newest running session is recommended without scores", func(t *testing.T) {
		now := fixedHistoryInsightsNow()
		older := historyInsightSession(21, 1, model.SessionStatusRunning, now.AddDate(0, 0, -3))
		newer := historyInsightSession(22, 2, model.SessionStatusRunning, now.AddDate(0, 0, -1))
		result := runHistoryInsightsRecommendationCase(t, now, []model.Session{newer, older}, newFakeFeedbackRepository(), newFakeReportRepository())

		if result.NextRecommendation == nil {
			t.Fatal("next recommendation is nil, want continue session recommendation")
		}
		if result.NextRecommendation.Type != "continue_session" {
			t.Fatalf("recommendation type = %q, want continue_session", result.NextRecommendation.Type)
		}
		if result.NextRecommendation.SessionID != newer.ID {
			t.Fatalf("recommendation session id = %d, want newest running session %d", result.NextRecommendation.SessionID, newer.ID)
		}
		if result.NextRecommendation.Scenario == nil || result.NextRecommendation.Scenario.ID != newer.ScenarioID {
			t.Fatalf("recommendation scenario = %+v, want scenario %d", result.NextRecommendation.Scenario, newer.ScenarioID)
		}
	})

	t.Run("nil without repeated errors scored scenarios or running sessions", func(t *testing.T) {
		now := fixedHistoryInsightsNow()
		result := runHistoryInsightsRecommendationCase(t, now, nil, newFakeFeedbackRepository(), newFakeReportRepository())

		if result.NextRecommendation != nil {
			t.Fatalf("next recommendation = %+v, want nil", result.NextRecommendation)
		}
	})
}

func TestHistoryInsightsRejectsInvalidRequest(t *testing.T) {
	now := fixedHistoryInsightsNow()
	tests := []struct {
		name  string
		input service.HistoryInsightsInput
	}{
		{
			name:  "unsupported days",
			input: service.HistoryInsightsInput{UserID: 1, Days: 14},
		},
		{
			name:  "negative user id",
			input: service.HistoryInsightsInput{UserID: -1, Days: 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := newFakeHistoryInsightSessionRepository()
			insightsService := newHistoryInsightsTestService(t, sessionRepo, nil, nil, nil, now)

			_, err := insightsService.GetInsights(tt.input)

			if !errors.Is(err, service.ErrInvalidHistoryRequest) {
				t.Fatalf("GetInsights error = %v, want ErrInvalidHistoryRequest", err)
			}
			if len(sessionRepo.queries) != 0 {
				t.Fatalf("repository queries = %+v, want none for invalid request", sessionRepo.queries)
			}
		})
	}
}

func runHistoryInsightsRecommendationCase(
	t *testing.T,
	now time.Time,
	currentSessions []model.Session,
	feedbackRepo *fakeFeedbackRepository,
	reportRepo *fakeReportRepository,
) service.HistoryInsightsResult {
	t.Helper()

	currentStart := now.AddDate(0, 0, -30)
	previousStart := now.AddDate(0, 0, -60)
	sessionRepo := newFakeHistoryInsightSessionRepository()
	sessionRepo.setWindow(currentStart, now, currentSessions)
	sessionRepo.setWindow(previousStart, currentStart, nil)
	insightsService := newHistoryInsightsTestService(t, sessionRepo, feedbackRepo, reportRepo, nil, now)
	result, err := insightsService.GetInsights(service.HistoryInsightsInput{UserID: 1, Days: 30})
	if err != nil {
		t.Fatalf("GetInsights returned error: %v", err)
	}

	return result
}

func newHistoryInsightsTestService(
	t *testing.T,
	sessionRepo *fakeHistoryInsightSessionRepository,
	feedbackRepo *fakeFeedbackRepository,
	reportRepo *fakeReportRepository,
	scenarios map[int]model.Scenario,
	now time.Time,
) *service.HistoryInsightsService {
	t.Helper()

	if feedbackRepo == nil {
		feedbackRepo = newFakeFeedbackRepository()
	}
	if reportRepo == nil {
		reportRepo = newFakeReportRepository()
	}
	if scenarios == nil {
		scenarios = map[int]model.Scenario{
			1: {ID: 1, Code: "interview", Name: "Interview", Difficulty: "medium"},
			2: {ID: 2, Code: "presentation", Name: "Presentation", Difficulty: "hard"},
		}
	}

	return service.NewHistoryInsightsService(
		fakeScenarioReader{scenarios: scenarios},
		sessionRepo,
		feedbackRepo,
		reportRepo,
		service.WithHistoryInsightsNow(func() time.Time { return now }),
		service.WithHistoryInsightsLocation(time.UTC),
	)
}

type fakeHistoryInsightSessionRepository struct {
	queries         []model.SessionWindowQuery
	sessionsByRange map[string][]model.Session
	err             error
}

func newFakeHistoryInsightSessionRepository() *fakeHistoryInsightSessionRepository {
	return &fakeHistoryInsightSessionRepository{
		sessionsByRange: make(map[string][]model.Session),
	}
}

func (r *fakeHistoryInsightSessionRepository) setWindow(start time.Time, end time.Time, sessions []model.Session) {
	r.sessionsByRange[historyInsightWindowKey(start, end)] = append([]model.Session(nil), sessions...)
}

func (r *fakeHistoryInsightSessionRepository) ListSessionsByWindow(query model.SessionWindowQuery) ([]model.Session, error) {
	r.queries = append(r.queries, query)
	if r.err != nil {
		return nil, r.err
	}

	sessions := r.sessionsByRange[historyInsightWindowKey(query.StartedAt, query.EndedAt)]
	return append([]model.Session(nil), sessions...), nil
}

func historyInsightWindowKey(start time.Time, end time.Time) string {
	return start.Format(time.RFC3339Nano) + "|" + end.Format(time.RFC3339Nano)
}

func historyInsightSession(id int, scenarioID int, status model.SessionStatus, createdAt time.Time) model.Session {
	return model.Session{
		ID:         id,
		SessionNo:  "SINSIGHT",
		ScenarioID: scenarioID,
		UserID:     1,
		Status:     status,
		TurnCount:  1,
		CreatedAt:  createdAt,
	}
}

func historyInsightScore(sessionID int, total int) model.ScoreResult {
	return model.ScoreResult{
		SessionID:  sessionID,
		TotalScore: total,
	}
}

func fixedHistoryInsightsNow() time.Time {
	return time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
}

func assertHistoryInsightIntPtr(t *testing.T, got *int, want int, name string) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s = nil, want %d", name, want)
	}
	if *got != want {
		t.Fatalf("%s = %d, want %d", name, *got, want)
	}
}

func assertHistoryInsightWindow(t *testing.T, got model.SessionWindowQuery, wantUserID int, wantStart time.Time, wantEnd time.Time) {
	t.Helper()

	if got.UserID != wantUserID {
		t.Fatalf("query user id = %d, want %d", got.UserID, wantUserID)
	}
	if !got.StartedAt.Equal(wantStart) {
		t.Fatalf("query started at = %s, want %s", got.StartedAt, wantStart)
	}
	if !got.EndedAt.Equal(wantEnd) {
		t.Fatalf("query ended at = %s, want %s", got.EndedAt, wantEnd)
	}
	if got.Limit != 500 {
		t.Fatalf("query limit = %d, want 500", got.Limit)
	}
}

func findHistoryScenarioTrend(t *testing.T, trends []service.ScenarioTrend, scenarioID int) service.ScenarioTrend {
	t.Helper()

	for _, trend := range trends {
		if trend.Scenario.ID == scenarioID {
			return trend
		}
	}

	t.Fatalf("scenario trend for scenario %d not found in %+v", scenarioID, trends)
	return service.ScenarioTrend{}
}

var _ service.HistoryInsightSessionRepository = (*fakeHistoryInsightSessionRepository)(nil)
var _ service.HistoryReportRepository = (*fakeReportRepository)(nil)
var _ service.ReportFeedbackReader = (*fakeFeedbackRepository)(nil)
