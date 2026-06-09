package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/service"
)

func TestHistoryInsightsHandlerGetsInsights(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	insightService := &fakeHistoryInsightsService{result: sampleHistoryInsightsResult()}
	handler := NewHistoryInsightsHandler(insightService)
	engine := gin.New()
	engine.GET("/api/v1/history/insights", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/insights?days=30&user_id=42", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if insightService.callCount != 1 {
		t.Fatalf("call count = %d, want 1", insightService.callCount)
	}
	if insightService.lastInput != (service.HistoryInsightsInput{Days: 30, UserID: 42}) {
		t.Fatalf("input = %+v, want days 30 user_id 42", insightService.lastInput)
	}

	var raw struct {
		Code    int                        `json:"code"`
		Message string                     `json:"message"`
		Data    map[string]json.RawMessage `json:"data"`
	}
	mustUnmarshalHistoryInsightsJSON(t, rec.Body.Bytes(), &raw)

	if raw.Code != 0 {
		t.Fatalf("code = %d, want 0", raw.Code)
	}
	if raw.Message != "success" {
		t.Fatalf("message = %q, want success", raw.Message)
	}
	for _, key := range []string{"summary", "score_trend", "scenario_trends", "frequent_errors", "next_recommendation"} {
		if _, ok := raw.Data[key]; !ok {
			t.Fatalf("data missing key %q; data = %s", key, string(rec.Body.Bytes()))
		}
	}

	var body historyInsightsHandlerResponse
	mustUnmarshalHistoryInsightsJSON(t, rec.Body.Bytes(), &body)
	if body.Data.Summary.Days != 30 {
		t.Fatalf("summary.days = %d, want 30", body.Data.Summary.Days)
	}
	if body.Data.Summary.AverageScore == nil || *body.Data.Summary.AverageScore != 82 {
		t.Fatalf("summary.average_score = %v, want 82", body.Data.Summary.AverageScore)
	}
	if body.Data.Summary.PreviousAverageScore != nil {
		t.Fatalf("summary.previous_average_score = %v, want nil", *body.Data.Summary.PreviousAverageScore)
	}
	if len(body.Data.ScoreTrend) != 1 {
		t.Fatalf("score_trend length = %d, want 1", len(body.Data.ScoreTrend))
	}
	if body.Data.ScoreTrend[0].AverageScore != 82 {
		t.Fatalf("score_trend[0].average_score = %d, want 82", body.Data.ScoreTrend[0].AverageScore)
	}
	if len(body.Data.ScenarioTrends) != 1 {
		t.Fatalf("scenario_trends length = %d, want 1", len(body.Data.ScenarioTrends))
	}
	trend := body.Data.ScenarioTrends[0]
	if trend.Scenario.Code != "interview" {
		t.Fatalf("scenario_trends[0].scenario.code = %q, want interview", trend.Scenario.Code)
	}
	if trend.ScoreDelta == nil || *trend.ScoreDelta != 10 {
		t.Fatalf("scenario_trends[0].score_delta = %v, want 10", trend.ScoreDelta)
	}
	assertHistoryInsightsRFC3339(t, trend.LastTrainedAt)
	if len(body.Data.FrequentErrors) != 1 {
		t.Fatalf("frequent_errors length = %d, want 1", len(body.Data.FrequentErrors))
	}
	if body.Data.FrequentErrors[0].SourceSessionID != 99 {
		t.Fatalf("frequent_errors[0].source_session_id = %d, want 99", body.Data.FrequentErrors[0].SourceSessionID)
	}
	assertHistoryInsightsRFC3339(t, body.Data.FrequentErrors[0].LastSeenAt)
	if body.Data.NextRecommendation == nil {
		t.Fatal("next_recommendation = nil, want recommendation")
	}
	if body.Data.NextRecommendation.Scenario == nil {
		t.Fatal("next_recommendation.scenario = nil, want scenario")
	}
	if body.Data.NextRecommendation.Scenario.Code != "interview" {
		t.Fatalf("next_recommendation.scenario.code = %q, want interview", body.Data.NextRecommendation.Scenario.Code)
	}
}

func TestHistoryInsightsHandlerPassesAbsentQueryValuesAsZero(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	insightService := &fakeHistoryInsightsService{}
	handler := NewHistoryInsightsHandler(insightService)
	engine := gin.New()
	engine.GET("/api/v1/history/insights", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/insights", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if insightService.lastInput != (service.HistoryInsightsInput{}) {
		t.Fatalf("input = %+v, want zero-value input", insightService.lastInput)
	}
}

func TestHistoryInsightsHandlerRejectsInvalidQueryValues(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "empty days", path: "/api/v1/history/insights?days=&user_id=42"},
		{name: "zero days", path: "/api/v1/history/insights?days=0&user_id=42"},
		{name: "negative days", path: "/api/v1/history/insights?days=-1&user_id=42"},
		{name: "non-numeric days", path: "/api/v1/history/insights?days=abc&user_id=42"},
		{name: "empty user id", path: "/api/v1/history/insights?days=30&user_id="},
		{name: "zero user id", path: "/api/v1/history/insights?days=30&user_id=0"},
		{name: "negative user id", path: "/api/v1/history/insights?days=30&user_id=-1"},
		{name: "non-numeric user id", path: "/api/v1/history/insights?days=30&user_id=abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.ReleaseMode)
			insightService := &fakeHistoryInsightsService{}
			handler := NewHistoryInsightsHandler(insightService)
			engine := gin.New()
			engine.GET("/api/v1/history/insights", handler.Get)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assertHistoryInsightsErrorResponse(t, rec, http.StatusBadRequest, 6001, "invalid history request")
			if insightService.callCount != 0 {
				t.Fatalf("call count = %d, want 0", insightService.callCount)
			}
		})
	}
}

func TestHistoryInsightsHandlerSerializesRecommendationScenarioNull(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	insightService := &fakeHistoryInsightsService{
		result: service.HistoryInsightsResult{
			NextRecommendation: &service.NextPracticeRecommendation{
				Type:      "continue_session",
				Reason:    "A recent practice session is still running.",
				SessionID: 7,
				Focus:     "英语面试",
			},
		},
	}
	handler := NewHistoryInsightsHandler(insightService)
	engine := gin.New()
	engine.GET("/api/v1/history/insights", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/insights", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var raw struct {
		Data struct {
			NextRecommendation struct {
				Scenario json.RawMessage `json:"scenario"`
			} `json:"next_recommendation"`
		} `json:"data"`
	}
	mustUnmarshalHistoryInsightsJSON(t, rec.Body.Bytes(), &raw)
	if string(raw.Data.NextRecommendation.Scenario) != "null" {
		t.Fatalf("next_recommendation.scenario = %s, want null", string(raw.Data.NextRecommendation.Scenario))
	}
}

func TestHistoryInsightsHandlerMapsInvalidServiceRequest(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	insightService := &fakeHistoryInsightsService{err: service.ErrInvalidHistoryRequest}
	handler := NewHistoryInsightsHandler(insightService)
	engine := gin.New()
	engine.GET("/api/v1/history/insights", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/insights?days=14", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertHistoryInsightsErrorResponse(t, rec, http.StatusBadRequest, 6001, "invalid history request")
	if insightService.lastInput != (service.HistoryInsightsInput{Days: 14}) {
		t.Fatalf("input = %+v, want days 14", insightService.lastInput)
	}
}

func TestHistoryInsightsHandlerMapsInternalServiceError(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	insightService := &fakeHistoryInsightsService{err: errors.New("store failed")}
	handler := NewHistoryInsightsHandler(insightService)
	engine := gin.New()
	engine.GET("/api/v1/history/insights", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/history/insights?days=30&user_id=42", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertHistoryInsightsErrorResponse(t, rec, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
}

type fakeHistoryInsightsService struct {
	result    service.HistoryInsightsResult
	err       error
	callCount int
	lastInput service.HistoryInsightsInput
}

func (s *fakeHistoryInsightsService) GetInsights(input service.HistoryInsightsInput) (service.HistoryInsightsResult, error) {
	s.callCount++
	s.lastInput = input
	if s.err != nil {
		return service.HistoryInsightsResult{}, s.err
	}

	return s.result, nil
}

type historyInsightsHandlerResponse struct {
	Code int `json:"code"`
	Data struct {
		Summary struct {
			Days                 int  `json:"days"`
			TotalSessions        int  `json:"total_sessions"`
			FinishedSessions     int  `json:"finished_sessions"`
			RunningSessions      int  `json:"running_sessions"`
			ScoredSessions       int  `json:"scored_sessions"`
			GeneratedReports     int  `json:"generated_reports"`
			AverageScore         *int `json:"average_score"`
			PreviousAverageScore *int `json:"previous_average_score"`
			ScoreDelta           *int `json:"score_delta"`
		} `json:"summary"`
		ScoreTrend []struct {
			Date         string `json:"date"`
			AverageScore int    `json:"average_score"`
			SessionCount int    `json:"session_count"`
		} `json:"score_trend"`
		ScenarioTrends []struct {
			Scenario       historyInsightsHandlerScenario `json:"scenario"`
			SessionCount   int                            `json:"session_count"`
			ScoredSessions int                            `json:"scored_sessions"`
			AverageScore   *int                           `json:"average_score"`
			FirstScore     *int                           `json:"first_score"`
			LatestScore    *int                           `json:"latest_score"`
			ScoreDelta     *int                           `json:"score_delta"`
			LastTrainedAt  string                         `json:"last_trained_at"`
		} `json:"scenario_trends"`
		FrequentErrors []struct {
			Key             string `json:"key"`
			Title           string `json:"title"`
			Category        string `json:"category"`
			Suggestion      string `json:"suggestion"`
			Count           int    `json:"count"`
			LatestEvidence  string `json:"latest_evidence"`
			LastSeenAt      string `json:"last_seen_at"`
			SourceSessionID int    `json:"source_session_id"`
		} `json:"frequent_errors"`
		NextRecommendation *struct {
			Type      string                          `json:"type"`
			Reason    string                          `json:"reason"`
			Scenario  *historyInsightsHandlerScenario `json:"scenario"`
			SessionID int                             `json:"session_id"`
			Focus     string                          `json:"focus"`
		} `json:"next_recommendation"`
	} `json:"data"`
}

type historyInsightsHandlerScenario struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
}

func sampleHistoryInsightsResult() service.HistoryInsightsResult {
	average := 82
	firstScore := 72
	latestScore := 82
	scoreDelta := 10
	scenario := model.Scenario{
		ID:          1,
		Code:        "interview",
		Name:        "英语面试",
		Description: "模拟技术面试回答。",
		Difficulty:  "medium",
	}

	return service.HistoryInsightsResult{
		Summary: service.HistoryInsightSummary{
			Days:             30,
			TotalSessions:    3,
			FinishedSessions: 2,
			RunningSessions:  1,
			ScoredSessions:   2,
			GeneratedReports: 1,
			AverageScore:     &average,
		},
		ScoreTrend: []service.HistoryScoreTrendPoint{
			{Date: "2026-06-09", AverageScore: 82, SessionCount: 2},
		},
		ScenarioTrends: []service.ScenarioTrend{
			{
				Scenario:       scenario,
				SessionCount:   2,
				ScoredSessions: 2,
				AverageScore:   &average,
				FirstScore:     &firstScore,
				LatestScore:    &latestScore,
				ScoreDelta:     &scoreDelta,
				LastTrainedAt:  time.Date(2026, 6, 9, 8, 30, 0, 0, time.FixedZone("CST", 8*60*60)),
			},
		},
		FrequentErrors: []service.FrequentErrorInsight{
			{
				Key:             "am study",
				Title:           "am study",
				Category:        "grammar",
				Suggestion:      "am studying",
				Count:           2,
				LatestEvidence:  "am study -> am studying",
				LastSeenAt:      time.Date(2026, 6, 9, 9, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
				SourceSessionID: 99,
			},
		},
		NextRecommendation: &service.NextPracticeRecommendation{
			Type:      "scenario_repractice",
			Reason:    "Repeated error appeared in recent reports.",
			Scenario:  &scenario,
			SessionID: 99,
			Focus:     "am study",
		},
	}
}

func assertHistoryInsightsErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, httpStatus int, code int, message string) {
	t.Helper()

	if rec.Code != httpStatus {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, httpStatus, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	mustUnmarshalHistoryInsightsJSON(t, rec.Body.Bytes(), &body)

	if body.Code != code {
		t.Fatalf("code = %d, want %d", body.Code, code)
	}
	if body.Message != message {
		t.Fatalf("message = %q, want %q", body.Message, message)
	}
}

func assertHistoryInsightsRFC3339(t *testing.T, value string) {
	t.Helper()

	if _, err := time.Parse(time.RFC3339, value); err != nil {
		t.Fatalf("time %q is not RFC3339: %v", value, err)
	}
}

func mustUnmarshalHistoryInsightsJSON(t *testing.T, data []byte, v any) {
	t.Helper()

	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}
