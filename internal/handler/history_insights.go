package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/response"
	"speakmate/internal/service"
)

// HistoryInsightsService defines the business capability used by the history insights handler.
type HistoryInsightsService interface {
	GetInsights(input service.HistoryInsightsInput) (service.HistoryInsightsResult, error)
}

// HistoryInsightsHandler handles training history insight APIs.
type HistoryInsightsHandler struct {
	service HistoryInsightsService
}

// NewHistoryInsightsHandler creates a history insights API handler.
func NewHistoryInsightsHandler(service HistoryInsightsService) *HistoryInsightsHandler {
	return &HistoryInsightsHandler{service: service}
}

// Get returns aggregated training history insights.
func (h *HistoryInsightsHandler) Get(c *gin.Context) {
	days, ok := parseHistoryInsightsPositiveQueryInt(c, "days")
	if !ok {
		return
	}
	userID, ok := parseHistoryInsightsPositiveQueryInt(c, "user_id")
	if !ok {
		return
	}

	result, err := h.service.GetInsights(service.HistoryInsightsInput{
		Days:   days,
		UserID: userID,
	})
	if err != nil {
		writeHistoryError(c, err)
		return
	}

	response.Success(c, toHistoryInsightsResponse(result))
}

func parseHistoryInsightsPositiveQueryInt(c *gin.Context, key string) (int, bool) {
	raw, exists := c.GetQuery(key)
	if !exists {
		return 0, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		response.Error(c, http.StatusBadRequest, invalidHistoryRequestCode, "历史记录请求无效")
		return 0, false
	}

	return value, true
}

type historyInsightsResponse struct {
	Summary            historyInsightSummaryResponse      `json:"summary"`
	ScoreTrend         []historyScoreTrendPointResponse   `json:"score_trend"`
	ScenarioTrends     []historyScenarioTrendResponse     `json:"scenario_trends"`
	FrequentErrors     []historyFrequentErrorResponse     `json:"frequent_errors"`
	NextRecommendation *historyNextRecommendationResponse `json:"next_recommendation"`
}

type historyInsightSummaryResponse struct {
	Days                 int  `json:"days"`
	TotalSessions        int  `json:"total_sessions"`
	FinishedSessions     int  `json:"finished_sessions"`
	RunningSessions      int  `json:"running_sessions"`
	ScoredSessions       int  `json:"scored_sessions"`
	GeneratedReports     int  `json:"generated_reports"`
	AverageScore         *int `json:"average_score"`
	PreviousAverageScore *int `json:"previous_average_score"`
	ScoreDelta           *int `json:"score_delta"`
}

type historyScoreTrendPointResponse struct {
	Date         string `json:"date"`
	AverageScore int    `json:"average_score"`
	SessionCount int    `json:"session_count"`
}

type historyScenarioTrendResponse struct {
	Scenario       scenarioSummary `json:"scenario"`
	SessionCount   int             `json:"session_count"`
	ScoredSessions int             `json:"scored_sessions"`
	AverageScore   *int            `json:"average_score"`
	FirstScore     *int            `json:"first_score"`
	LatestScore    *int            `json:"latest_score"`
	ScoreDelta     *int            `json:"score_delta"`
	LastTrainedAt  string          `json:"last_trained_at"`
}

type historyFrequentErrorResponse struct {
	Key             string `json:"key"`
	Title           string `json:"title"`
	Category        string `json:"category"`
	Suggestion      string `json:"suggestion"`
	Count           int    `json:"count"`
	LatestEvidence  string `json:"latest_evidence"`
	LastSeenAt      string `json:"last_seen_at"`
	SourceSessionID int    `json:"source_session_id"`
}

type historyNextRecommendationResponse struct {
	Type      string           `json:"type"`
	Reason    string           `json:"reason"`
	Scenario  *scenarioSummary `json:"scenario"`
	SessionID int              `json:"session_id"`
	Focus     string           `json:"focus"`
}

func toHistoryInsightsResponse(result service.HistoryInsightsResult) historyInsightsResponse {
	scoreTrend := make([]historyScoreTrendPointResponse, 0, len(result.ScoreTrend))
	for _, point := range result.ScoreTrend {
		scoreTrend = append(scoreTrend, historyScoreTrendPointResponse{
			Date:         point.Date,
			AverageScore: point.AverageScore,
			SessionCount: point.SessionCount,
		})
	}

	scenarioTrends := make([]historyScenarioTrendResponse, 0, len(result.ScenarioTrends))
	for _, trend := range result.ScenarioTrends {
		scenarioTrends = append(scenarioTrends, historyScenarioTrendResponse{
			Scenario:       toScenarioSummary(trend.Scenario),
			SessionCount:   trend.SessionCount,
			ScoredSessions: trend.ScoredSessions,
			AverageScore:   trend.AverageScore,
			FirstScore:     trend.FirstScore,
			LatestScore:    trend.LatestScore,
			ScoreDelta:     trend.ScoreDelta,
			LastTrainedAt:  formatTime(trend.LastTrainedAt),
		})
	}

	frequentErrors := make([]historyFrequentErrorResponse, 0, len(result.FrequentErrors))
	for _, insight := range result.FrequentErrors {
		frequentErrors = append(frequentErrors, historyFrequentErrorResponse{
			Key:             insight.Key,
			Title:           insight.Title,
			Category:        insight.Category,
			Suggestion:      insight.Suggestion,
			Count:           insight.Count,
			LatestEvidence:  insight.LatestEvidence,
			LastSeenAt:      formatTime(insight.LastSeenAt),
			SourceSessionID: insight.SourceSessionID,
		})
	}

	return historyInsightsResponse{
		Summary: historyInsightSummaryResponse{
			Days:                 result.Summary.Days,
			TotalSessions:        result.Summary.TotalSessions,
			FinishedSessions:     result.Summary.FinishedSessions,
			RunningSessions:      result.Summary.RunningSessions,
			ScoredSessions:       result.Summary.ScoredSessions,
			GeneratedReports:     result.Summary.GeneratedReports,
			AverageScore:         result.Summary.AverageScore,
			PreviousAverageScore: result.Summary.PreviousAverageScore,
			ScoreDelta:           result.Summary.ScoreDelta,
		},
		ScoreTrend:         scoreTrend,
		ScenarioTrends:     scenarioTrends,
		FrequentErrors:     frequentErrors,
		NextRecommendation: toHistoryNextRecommendationResponse(result.NextRecommendation),
	}
}

func toHistoryNextRecommendationResponse(recommendation *service.NextPracticeRecommendation) *historyNextRecommendationResponse {
	if recommendation == nil {
		return nil
	}

	var scenario *scenarioSummary
	if recommendation.Scenario != nil {
		summary := toScenarioSummary(*recommendation.Scenario)
		scenario = &summary
	}

	return &historyNextRecommendationResponse{
		Type:      recommendation.Type,
		Reason:    recommendation.Reason,
		Scenario:  scenario,
		SessionID: recommendation.SessionID,
		Focus:     recommendation.Focus,
	}
}

func toScenarioSummary(scenario model.Scenario) scenarioSummary {
	return scenarioSummary{
		ID:          scenario.ID,
		Code:        scenario.Code,
		Name:        scenario.Name,
		Description: scenario.Description,
		Difficulty:  scenario.Difficulty,
	}
}
