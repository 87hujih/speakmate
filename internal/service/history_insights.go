package service

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

const maxHistoryInsightsSessions = 500

type HistoryInsightsInput struct {
	Days   int
	UserID int
}

type HistoryInsightSessionRepository interface {
	ListSessionsByWindow(query model.SessionWindowQuery) ([]model.Session, error)
}

type HistoryInsightsService struct {
	scenarioReader ScenarioReader
	sessionRepo    HistoryInsightSessionRepository
	feedbackRepo   ReportFeedbackReader
	reportRepo     HistoryReportRepository
	now            func() time.Time
	location       *time.Location
}

type HistoryInsightsOption func(*HistoryInsightsService)

func NewHistoryInsightsService(
	scenarioReader ScenarioReader,
	sessionRepo HistoryInsightSessionRepository,
	feedbackRepo ReportFeedbackReader,
	reportRepo HistoryReportRepository,
	opts ...HistoryInsightsOption,
) *HistoryInsightsService {
	service := &HistoryInsightsService{
		scenarioReader: scenarioReader,
		sessionRepo:    sessionRepo,
		feedbackRepo:   feedbackRepo,
		reportRepo:     reportRepo,
		now:            time.Now,
		location:       time.Local,
	}
	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithHistoryInsightsNow(now func() time.Time) HistoryInsightsOption {
	return func(service *HistoryInsightsService) {
		if now != nil {
			service.now = now
		}
	}
}

func WithHistoryInsightsLocation(location *time.Location) HistoryInsightsOption {
	return func(service *HistoryInsightsService) {
		if location != nil {
			service.location = location
		}
	}
}

type HistoryInsightsResult struct {
	Summary            HistoryInsightSummary
	ScoreTrend         []HistoryScoreTrendPoint
	ScenarioTrends     []ScenarioTrend
	FrequentErrors     []FrequentErrorInsight
	NextRecommendation *NextPracticeRecommendation
}

type HistoryInsightSummary struct {
	Days                 int
	TotalSessions        int
	FinishedSessions     int
	RunningSessions      int
	ScoredSessions       int
	GeneratedReports     int
	AverageScore         *int
	PreviousAverageScore *int
	ScoreDelta           *int
}

type HistoryScoreTrendPoint struct {
	Date         string
	AverageScore int
	SessionCount int
}

type ScenarioTrend struct {
	Scenario       model.Scenario
	SessionCount   int
	ScoredSessions int
	AverageScore   *int
	FirstScore     *int
	LatestScore    *int
	ScoreDelta     *int
	LastTrainedAt  time.Time
}

type FrequentErrorInsight struct {
	Key             string
	Title           string
	Category        string
	Suggestion      string
	Count           int
	LatestEvidence  string
	LastSeenAt      time.Time
	SourceSessionID int
}

type NextPracticeRecommendation struct {
	Type      string
	Reason    string
	Scenario  *model.Scenario
	SessionID int
	Focus     string
}

func (s *HistoryInsightsService) GetInsights(input HistoryInsightsInput) (HistoryInsightsResult, error) {
	days, err := normalizeHistoryInsightsDays(input.Days)
	if err != nil {
		return HistoryInsightsResult{}, err
	}
	if input.UserID < 0 {
		return HistoryInsightsResult{}, ErrInvalidHistoryRequest
	}

	now := s.historyInsightsNow()
	currentStart := now.AddDate(0, 0, -days)
	previousStart := now.AddDate(0, 0, -days*2)
	currentSessions, err := s.sessionRepo.ListSessionsByWindow(model.SessionWindowQuery{
		UserID:    input.UserID,
		StartedAt: currentStart,
		EndedAt:   now,
		Limit:     maxHistoryInsightsSessions,
	})
	if err != nil {
		return HistoryInsightsResult{}, err
	}
	previousSessions, err := s.sessionRepo.ListSessionsByWindow(model.SessionWindowQuery{
		UserID:    input.UserID,
		StartedAt: previousStart,
		EndedAt:   currentStart,
		Limit:     maxHistoryInsightsSessions,
	})
	if err != nil {
		return HistoryInsightsResult{}, err
	}

	currentScores, err := s.scoreBySession(currentSessions)
	if err != nil {
		return HistoryInsightsResult{}, err
	}
	previousScores, err := s.scoreBySession(previousSessions)
	if err != nil {
		return HistoryInsightsResult{}, err
	}
	scenarioCache := map[int]model.Scenario{}
	scenarioLookup := func(id int) (model.Scenario, error) {
		return s.scenarioByID(scenarioCache, id)
	}
	scoreTrend := buildScoreTrend(currentSessions, currentScores, s.historyInsightsLocation())
	scenarioTrends, err := buildScenarioTrends(currentSessions, currentScores, scenarioLookup)
	if err != nil {
		return HistoryInsightsResult{}, err
	}
	frequentErrors, generatedReports, err := s.buildFrequentErrors(currentSessions)
	if err != nil {
		return HistoryInsightsResult{}, err
	}
	recommendation, err := s.buildNextRecommendation(currentSessions, currentScores, scenarioTrends, frequentErrors, scenarioLookup)
	if err != nil {
		return HistoryInsightsResult{}, err
	}

	currentAverage := averageRounded(finishedScores(currentSessions, currentScores))
	previousAverage := averageRounded(finishedScores(previousSessions, previousScores))
	summary := buildHistoryInsightSummary(days, currentSessions, currentScores, generatedReports, currentAverage, previousAverage)

	return HistoryInsightsResult{
		Summary:            summary,
		ScoreTrend:         scoreTrend,
		ScenarioTrends:     scenarioTrends,
		FrequentErrors:     frequentErrors,
		NextRecommendation: recommendation,
	}, nil
}

func normalizeHistoryInsightsDays(days int) (int, error) {
	if days == 0 {
		return 30, nil
	}
	if days != 7 && days != 30 {
		return 0, ErrInvalidHistoryRequest
	}

	return days, nil
}

func (s *HistoryInsightsService) historyInsightsNow() time.Time {
	if s.now == nil {
		return time.Now()
	}

	return s.now()
}

func (s *HistoryInsightsService) historyInsightsLocation() *time.Location {
	if s.location == nil {
		return time.Local
	}

	return s.location
}

func (s *HistoryInsightsService) scoreBySession(sessions []model.Session) (map[int]model.ScoreResult, error) {
	scores := map[int]model.ScoreResult{}
	if s.feedbackRepo == nil {
		return scores, nil
	}

	for _, session := range sessions {
		score, err := s.feedbackRepo.FindCurrentScoreBySessionID(session.ID)
		if err != nil {
			if isHistoryInsightsScoreNotFound(err) {
				continue
			}

			return nil, err
		}
		scores[session.ID] = score
	}

	return scores, nil
}

func (s *HistoryInsightsService) scenarioByID(cache map[int]model.Scenario, id int) (model.Scenario, error) {
	if scenario, ok := cache[id]; ok {
		return scenario, nil
	}
	scenario, err := s.scenarioReader.GetScenario(id)
	if err != nil {
		return model.Scenario{}, err
	}
	cache[id] = scenario

	return scenario, nil
}

func buildHistoryInsightSummary(
	days int,
	sessions []model.Session,
	scores map[int]model.ScoreResult,
	generatedReports int,
	currentAverage *int,
	previousAverage *int,
) HistoryInsightSummary {
	summary := HistoryInsightSummary{
		Days:                 days,
		TotalSessions:        len(sessions),
		GeneratedReports:     generatedReports,
		AverageScore:         currentAverage,
		PreviousAverageScore: previousAverage,
	}
	for _, session := range sessions {
		switch session.Status {
		case model.SessionStatusFinished:
			summary.FinishedSessions++
			if _, ok := scores[session.ID]; ok {
				summary.ScoredSessions++
			}
		case model.SessionStatusRunning:
			summary.RunningSessions++
		}
	}
	if currentAverage != nil && previousAverage != nil {
		summary.ScoreDelta = intPtr(*currentAverage - *previousAverage)
	}

	return summary
}

func finishedScores(sessions []model.Session, scores map[int]model.ScoreResult) []int {
	values := []int{}
	for _, session := range sessions {
		if session.Status != model.SessionStatusFinished {
			continue
		}
		score, ok := scores[session.ID]
		if !ok {
			continue
		}
		values = append(values, score.TotalScore)
	}

	return values
}

func averageRounded(values []int) *int {
	if len(values) == 0 {
		return nil
	}
	total := 0
	for _, value := range values {
		total += value
	}

	return intPtr(int(math.Round(float64(total) / float64(len(values)))))
}

func buildScoreTrend(sessions []model.Session, scores map[int]model.ScoreResult, location *time.Location) []HistoryScoreTrendPoint {
	buckets := map[string][]int{}
	for _, session := range sessions {
		if session.Status != model.SessionStatusFinished {
			continue
		}
		score, ok := scores[session.ID]
		if !ok {
			continue
		}
		date := session.CreatedAt.In(location).Format("2006-01-02")
		buckets[date] = append(buckets[date], score.TotalScore)
	}

	dates := make([]string, 0, len(buckets))
	for date := range buckets {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	points := make([]HistoryScoreTrendPoint, 0, len(dates))
	for _, date := range dates {
		average := averageRounded(buckets[date])
		points = append(points, HistoryScoreTrendPoint{
			Date:         date,
			AverageScore: *average,
			SessionCount: len(buckets[date]),
		})
	}

	return points
}

type scenarioTrendBucket struct {
	scenarioID     int
	sessionCount   int
	scoredSessions int
	scores         []int
	firstScore     *int
	firstScoreAt   time.Time
	latestScore    *int
	latestScoreAt  time.Time
	lastTrainedAt  time.Time
}

func buildScenarioTrends(
	sessions []model.Session,
	scores map[int]model.ScoreResult,
	scenarioLookup func(int) (model.Scenario, error),
) ([]ScenarioTrend, error) {
	buckets := map[int]*scenarioTrendBucket{}
	for _, session := range sessions {
		bucket, ok := buckets[session.ScenarioID]
		if !ok {
			bucket = &scenarioTrendBucket{scenarioID: session.ScenarioID}
			buckets[session.ScenarioID] = bucket
		}
		bucket.sessionCount++
		if bucket.lastTrainedAt.IsZero() || session.CreatedAt.After(bucket.lastTrainedAt) {
			bucket.lastTrainedAt = session.CreatedAt
		}
		if session.Status != model.SessionStatusFinished {
			continue
		}
		score, ok := scores[session.ID]
		if !ok {
			continue
		}
		bucket.scoredSessions++
		bucket.scores = append(bucket.scores, score.TotalScore)
		if bucket.firstScore == nil || session.CreatedAt.Before(bucket.firstScoreAt) {
			bucket.firstScore = intPtr(score.TotalScore)
			bucket.firstScoreAt = session.CreatedAt
		}
		if bucket.latestScore == nil || session.CreatedAt.After(bucket.latestScoreAt) {
			bucket.latestScore = intPtr(score.TotalScore)
			bucket.latestScoreAt = session.CreatedAt
		}
	}

	trends := make([]ScenarioTrend, 0, len(buckets))
	for _, bucket := range buckets {
		scenario, err := scenarioLookup(bucket.scenarioID)
		if err != nil {
			return nil, err
		}
		trend := ScenarioTrend{
			Scenario:       scenario,
			SessionCount:   bucket.sessionCount,
			ScoredSessions: bucket.scoredSessions,
			AverageScore:   averageRounded(bucket.scores),
			FirstScore:     cloneIntPtr(bucket.firstScore),
			LatestScore:    cloneIntPtr(bucket.latestScore),
			LastTrainedAt:  bucket.lastTrainedAt,
		}
		if trend.FirstScore != nil && trend.LatestScore != nil {
			trend.ScoreDelta = intPtr(*trend.LatestScore - *trend.FirstScore)
		}
		trends = append(trends, trend)
	}

	sort.Slice(trends, func(i int, j int) bool {
		if trends[i].LastTrainedAt.Equal(trends[j].LastTrainedAt) {
			return trends[i].Scenario.ID < trends[j].Scenario.ID
		}

		return trends[i].LastTrainedAt.After(trends[j].LastTrainedAt)
	})

	return trends, nil
}

func (s *HistoryInsightsService) buildFrequentErrors(sessions []model.Session) ([]FrequentErrorInsight, int, error) {
	if s.reportRepo == nil {
		return []FrequentErrorInsight{}, 0, nil
	}

	errorsByKey := map[string]FrequentErrorInsight{}
	generatedReports := 0
	for _, session := range sessions {
		report, err := s.reportRepo.FindBySessionID(session.ID)
		if err != nil {
			if isHistoryInsightsReportNotFound(err) {
				continue
			}

			return nil, 0, err
		}
		generatedReports++
		lastSeenAt := report.CreatedAt
		if lastSeenAt.IsZero() {
			lastSeenAt = session.CreatedAt
		}
		for _, raw := range report.FrequentErrors {
			parsed, ok := parseFrequentErrorInsight(raw)
			if !ok {
				continue
			}
			existing := errorsByKey[parsed.Key]
			parsed.Count = existing.Count + 1
			if existing.Key != "" && !lastSeenAt.After(existing.LastSeenAt) {
				parsed.Title = existing.Title
				parsed.Category = existing.Category
				parsed.Suggestion = existing.Suggestion
				parsed.LatestEvidence = existing.LatestEvidence
				parsed.LastSeenAt = existing.LastSeenAt
				parsed.SourceSessionID = existing.SourceSessionID
			} else {
				parsed.LastSeenAt = lastSeenAt
				parsed.SourceSessionID = session.ID
			}
			errorsByKey[parsed.Key] = parsed
		}
	}

	frequentErrors := make([]FrequentErrorInsight, 0, len(errorsByKey))
	for _, insight := range errorsByKey {
		frequentErrors = append(frequentErrors, insight)
	}
	sort.Slice(frequentErrors, func(i int, j int) bool {
		if frequentErrors[i].Count != frequentErrors[j].Count {
			return frequentErrors[i].Count > frequentErrors[j].Count
		}
		if !frequentErrors[i].LastSeenAt.Equal(frequentErrors[j].LastSeenAt) {
			return frequentErrors[i].LastSeenAt.After(frequentErrors[j].LastSeenAt)
		}

		return frequentErrors[i].Key < frequentErrors[j].Key
	})
	if len(frequentErrors) > 5 {
		frequentErrors = frequentErrors[:5]
	}

	return frequentErrors, generatedReports, nil
}

func parseFrequentErrorInsight(raw string) (FrequentErrorInsight, bool) {
	evidence := strings.TrimSpace(raw)
	if evidence == "" {
		return FrequentErrorInsight{}, false
	}
	segments := strings.Split(evidence, "|")
	firstSegment := strings.TrimSpace(segments[0])
	if firstSegment == "" || isNoFrequentErrorPlaceholder(firstSegment) {
		return FrequentErrorInsight{}, false
	}

	title := firstSegment
	suggestion := ""
	parts := strings.SplitN(firstSegment, "->", 2)
	if len(parts) == 2 {
		title = strings.TrimSpace(parts[0])
		suggestion = strings.TrimSpace(parts[1])
	}
	key := strings.ToLower(strings.TrimSpace(title))
	if key == "" {
		return FrequentErrorInsight{}, false
	}

	return FrequentErrorInsight{
		Key:            key,
		Title:          title,
		Category:       "grammar",
		Suggestion:     suggestion,
		LatestEvidence: evidence,
	}, true
}

func isNoFrequentErrorPlaceholder(value string) bool {
	trimmed := strings.TrimSpace(value)
	lower := strings.ToLower(trimmed)

	return strings.HasPrefix(trimmed, "暂未发现高频错误") ||
		lower == "none" ||
		lower == "n/a" ||
		strings.HasPrefix(lower, "no frequent error")
}

func (s *HistoryInsightsService) buildNextRecommendation(
	sessions []model.Session,
	scores map[int]model.ScoreResult,
	scenarioTrends []ScenarioTrend,
	frequentErrors []FrequentErrorInsight,
	scenarioLookup func(int) (model.Scenario, error),
) (*NextPracticeRecommendation, error) {
	sessionByID := map[int]model.Session{}
	for _, session := range sessions {
		sessionByID[session.ID] = session
	}
	for _, frequentError := range frequentErrors {
		if frequentError.Count < 2 {
			continue
		}
		recommendation := NextPracticeRecommendation{
			Type:      "scenario_repractice",
			Reason:    "Repeated error appeared in recent reports.",
			SessionID: frequentError.SourceSessionID,
			Focus:     frequentError.Key,
		}
		if session, ok := sessionByID[frequentError.SourceSessionID]; ok {
			scenario, err := scenarioLookup(session.ScenarioID)
			if err != nil {
				return nil, err
			}
			recommendation.Scenario = &scenario
		}

		return &recommendation, nil
	}

	if trend := weakestScenarioTrend(scenarioTrends); trend != nil {
		scenario := trend.Scenario
		return &NextPracticeRecommendation{
			Type:      "scenario_repractice",
			Reason:    "This scenario has the weakest recent score.",
			Scenario:  &scenario,
			SessionID: latestScoredSessionIDForScenario(sessions, scores, scenario.ID),
			Focus:     scenario.Name,
		}, nil
	}

	if running := newestRunningSession(sessions); running != nil {
		scenario, err := scenarioLookup(running.ScenarioID)
		if err != nil {
			return nil, err
		}

		return &NextPracticeRecommendation{
			Type:      "continue_session",
			Reason:    "A recent practice session is still running.",
			Scenario:  &scenario,
			SessionID: running.ID,
			Focus:     scenario.Name,
		}, nil
	}

	return nil, nil
}

func weakestScenarioTrend(trends []ScenarioTrend) *ScenarioTrend {
	var weakest *ScenarioTrend
	for i := range trends {
		if trends[i].AverageScore == nil {
			continue
		}
		if weakest == nil ||
			*trends[i].AverageScore < *weakest.AverageScore ||
			(*trends[i].AverageScore == *weakest.AverageScore && trends[i].LastTrainedAt.After(weakest.LastTrainedAt)) {
			weakest = &trends[i]
		}
	}

	return weakest
}

func latestScoredSessionIDForScenario(sessions []model.Session, scores map[int]model.ScoreResult, scenarioID int) int {
	var latest model.Session
	found := false
	for _, session := range sessions {
		if session.ScenarioID != scenarioID || session.Status != model.SessionStatusFinished {
			continue
		}
		if _, ok := scores[session.ID]; !ok {
			continue
		}
		if !found || session.CreatedAt.After(latest.CreatedAt) || (session.CreatedAt.Equal(latest.CreatedAt) && session.ID > latest.ID) {
			latest = session
			found = true
		}
	}
	if !found {
		return 0
	}

	return latest.ID
}

func newestRunningSession(sessions []model.Session) *model.Session {
	var newest model.Session
	found := false
	for _, session := range sessions {
		if session.Status != model.SessionStatusRunning {
			continue
		}
		if !found || session.CreatedAt.After(newest.CreatedAt) || (session.CreatedAt.Equal(newest.CreatedAt) && session.ID > newest.ID) {
			newest = session
			found = true
		}
	}
	if !found {
		return nil
	}

	return &newest
}

func isHistoryInsightsScoreNotFound(err error) bool {
	return errors.Is(err, repository.ErrScoreNotFound) || errors.Is(err, ErrScoreNotFound)
}

func isHistoryInsightsReportNotFound(err error) bool {
	return errors.Is(err, repository.ErrReportNotFound) || errors.Is(err, ErrReportNotFound)
}

func intPtr(value int) *int {
	return &value
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}

	return intPtr(*value)
}
