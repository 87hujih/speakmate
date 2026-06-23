package service

import (
	"errors"
	"fmt"
	"time"

	"speakmate/internal/agent"
	"speakmate/internal/model"
	"speakmate/internal/repository"
	"speakmate/internal/stream"
)

// 服务层复用的哨兵错误。
var (
	// ErrInvalidReportRequest 表示报告请求参数非法。
	ErrInvalidReportRequest = errors.New("报告请求无效")
	// ErrSessionNotFinished 表示 Session 尚未结束，不能生成报告。
	ErrSessionNotFinished = errors.New("训练尚未结束")
	// ErrReportNotFound 表示业务层没有找到对应报告。
	ErrReportNotFound = errors.New("未找到课后报告")
	// ErrReportFeedbackMissing 表示报告生成缺少纠错或评分数据。
	ErrReportFeedbackMissing = errors.New("报告缺少反馈数据")
	// ErrSummaryAgentFailed 表示 Summary Agent 生成报告内容失败。
	ErrSummaryAgentFailed = errors.New("报告摘要生成失败")
)

// ReportSessionReader 定义报告服务依赖的 Session 读取能力。
type ReportSessionReader interface {
	FindByID(id int) (model.Session, error)
}

// ReportFeedbackReader 定义报告服务依赖的反馈读取能力。
type ReportFeedbackReader interface {
	ListCorrectionsBySessionID(sessionID int) ([]model.CorrectionResult, error)
	FindCurrentScoreBySessionID(sessionID int) (model.ScoreResult, error)
}

// ReportRepository 定义报告服务依赖的数据访问能力。
type ReportRepository interface {
	Save(report model.Report) error
	FindBySessionID(sessionID int) (model.Report, error)
}

// ReportService 封装课后报告生成和查询业务流程。
type ReportService struct {
	scenarioReader ScenarioReader
	sessionRepo    ReportSessionReader
	feedbackRepo   ReportFeedbackReader
	reportRepo     ReportRepository
	summary        agent.SummaryAgent
	events         EventPublisher
	now            func() time.Time
}

// ReportOption 用于配置 ReportService。
type ReportOption func(*ReportService)

// NewReportService 创建 Report 服务实例。
func NewReportService(
	scenarioReader ScenarioReader,
	sessionRepo ReportSessionReader,
	feedbackRepo ReportFeedbackReader,
	reportRepo ReportRepository,
	opts ...ReportOption,
) *ReportService {
	service := &ReportService{
		scenarioReader: scenarioReader,
		sessionRepo:    sessionRepo,
		feedbackRepo:   feedbackRepo,
		reportRepo:     reportRepo,
		summary:        agent.NewMockSummaryAgent(),
		now:            time.Now,
	}
	for _, opt := range opts {
		opt(service)
	}

	return service
}

// WithSummaryAgent 返回用于覆盖默认行为的配置选项。
func WithSummaryAgent(summary agent.SummaryAgent) ReportOption {
	return func(service *ReportService) {
		if summary != nil {
			service.summary = summary
		}
	}
}

// WithReportNow 返回用于覆盖默认行为的配置选项。
func WithReportNow(now func() time.Time) ReportOption {
	return func(service *ReportService) {
		if now != nil {
			service.now = now
		}
	}
}

// WithReportEventPublisher 返回用于覆盖默认行为的配置选项。
func WithReportEventPublisher(publisher EventPublisher) ReportOption {
	return func(service *ReportService) {
		if publisher != nil {
			service.events = publisher
		}
	}
}

// GenerateReport 基于已结束训练和反馈数据生成并保存课后报告。
func (s *ReportService) GenerateReport(sessionID int) (model.Report, error) {
	if sessionID <= 0 {
		return model.Report{}, ErrInvalidReportRequest
	}

	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return model.Report{}, s.publishReportError(sessionID, ErrSessionNotFound)
		}

		return model.Report{}, s.publishReportError(sessionID, err)
	}
	if session.Status != model.SessionStatusFinished {
		return model.Report{}, s.publishReportError(sessionID, ErrSessionNotFinished)
	}

	scenario, err := s.scenarioReader.GetScenario(session.ScenarioID)
	if err != nil {
		return model.Report{}, s.publishReportError(sessionID, err)
	}

	corrections, err := s.feedbackRepo.ListCorrectionsBySessionID(session.ID)
	if err != nil {
		if errors.Is(err, repository.ErrCorrectionNotFound) || errors.Is(err, ErrCorrectionNotFound) {
			return model.Report{}, s.publishReportError(sessionID, ErrReportFeedbackMissing)
		}

		return model.Report{}, s.publishReportError(sessionID, err)
	}
	score, err := s.feedbackRepo.FindCurrentScoreBySessionID(session.ID)
	if err != nil {
		if errors.Is(err, repository.ErrScoreNotFound) || errors.Is(err, ErrScoreNotFound) {
			return model.Report{}, s.publishReportError(sessionID, ErrReportFeedbackMissing)
		}

		return model.Report{}, s.publishReportError(sessionID, err)
	}

	summaryAgent := s.summary
	if summaryAgent == nil {
		summaryAgent = agent.NewMockSummaryAgent()
	}
	summaryOutput, err := summaryAgent.Summarize(agent.SummaryInput{
		Scenario:    scenario,
		Session:     session,
		Messages:    session.Messages,
		Corrections: corrections,
		Score:       score,
	})
	if err != nil {
		return model.Report{}, s.publishReportError(sessionID, fmt.Errorf("%w: %v", ErrSummaryAgentFailed, err))
	}

	report := model.Report{
		SessionID: session.ID,
		Scenario: model.ReportScenario{
			ID:         scenario.ID,
			Code:       scenario.Code,
			Name:       scenario.Name,
			Difficulty: scenario.Difficulty,
		},
		DurationSeconds:   durationSeconds(session),
		TurnCount:         session.TurnCount,
		TotalScore:        score.TotalScore,
		Scores:            score,
		Summary:           summaryOutput.Summary,
		MajorProblems:     stringListOrEmpty(summaryOutput.MajorProblems),
		FrequentErrors:    stringListOrEmpty(summaryOutput.FrequentErrors),
		BetterExpressions: stringListOrEmpty(summaryOutput.BetterExpressions),
		NextPracticePlan:  stringListOrEmpty(summaryOutput.NextPracticePlan),
		CreatedAt:         s.now(),
	}
	if err := s.reportRepo.Save(report); err != nil {
		return model.Report{}, s.publishReportError(sessionID, err)
	}
	if err := s.publishReportDone(report); err != nil {
		return model.Report{}, err
	}

	return report, nil
}

// GetReport 查询已生成的课后报告。
func (s *ReportService) GetReport(sessionID int) (model.Report, error) {
	if sessionID <= 0 {
		return model.Report{}, ErrInvalidReportRequest
	}

	report, err := s.reportRepo.FindBySessionID(sessionID)
	if err == nil {
		return report, nil
	}
	if errors.Is(err, repository.ErrReportNotFound) {
		return model.Report{}, ErrReportNotFound
	}

	return model.Report{}, err
}

// durationSeconds 计算训练持续秒数。
func durationSeconds(session model.Session) int {
	if session.EndedAt == nil || session.CreatedAt.IsZero() {
		return 0
	}
	duration := session.EndedAt.Sub(session.CreatedAt)
	if duration <= 0 {
		return 0
	}

	return int(duration.Seconds())
}

// stringListOrEmpty 将空字符串数组替换为默认列表。
func stringListOrEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}

	return append([]string(nil), values...)
}

// publishReportDone 发布报告生成完成事件。
func (s *ReportService) publishReportDone(report model.Report) error {
	if s.events == nil {
		return nil
	}

	if err := s.events.Publish(stream.Event{
		Type:      stream.EventTypeReportDone,
		SessionID: report.SessionID,
		Payload: stream.ReportDonePayload{
			TotalScore: report.TotalScore,
			Summary:    report.Summary,
		},
		CreatedAt: s.now(),
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrEventPublishFailed, err)
	}

	return nil
}

// publishReportError 发布报告生成失败事件。
func (s *ReportService) publishReportError(sessionID int, err error) error {
	if s.events == nil {
		return err
	}

	code, message := reportErrorPayload(err)
	_ = s.events.Publish(stream.Event{
		Type:      stream.EventTypeError,
		SessionID: sessionID,
		Payload: stream.ErrorPayload{
			Code:    code,
			Message: message,
		},
		CreatedAt: s.now(),
	})

	return err
}

// reportErrorPayload 将报告错误转换为 SSE 错误载荷。
func reportErrorPayload(err error) (string, string) {
	switch {
	case errors.Is(err, ErrSessionNotFound):
		return "session_not_found", "未找到训练"
	case errors.Is(err, ErrSessionNotFinished):
		return "session_not_finished", "训练尚未结束"
	case errors.Is(err, ErrReportFeedbackMissing):
		return "report_feedback_missing", "报告缺少反馈数据"
	case errors.Is(err, ErrSummaryAgentFailed):
		return "summary_agent_failed", "报告摘要生成失败"
	default:
		return "report_generation_failed", "报告生成失败"
	}
}
