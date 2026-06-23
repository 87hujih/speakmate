package service

import (
	"errors"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

// 当前服务使用的默认值和业务常量。
const (
	defaultHistoryPage     = 1
	defaultHistoryPageSize = 20
	maxHistoryPageSize     = 100
)

// ErrInvalidHistoryRequest 表示历史记录查询参数非法。
var ErrInvalidHistoryRequest = errors.New("历史记录请求无效")

// HistorySessionRepository 定义历史记录服务依赖的 Session 列表能力。
type HistorySessionRepository interface {
	ListSessions(query model.SessionListQuery) (model.SessionListResult, error)
}

// HistoryReportRepository 定义历史记录服务依赖的报告读取能力。
type HistoryReportRepository interface {
	FindBySessionID(sessionID int) (model.Report, error)
}

// HistoryService 封装训练历史查询业务流程。
type HistoryService struct {
	scenarioReader ScenarioReader
	sessionRepo    HistorySessionRepository
	feedbackRepo   ReportFeedbackReader
	reportRepo     HistoryReportRepository
}

// NewHistoryService 创建 History 服务实例。
func NewHistoryService(
	scenarioReader ScenarioReader,
	sessionRepo HistorySessionRepository,
	feedbackRepo ReportFeedbackReader,
	reportRepo HistoryReportRepository,
) *HistoryService {
	return &HistoryService{
		scenarioReader: scenarioReader,
		sessionRepo:    sessionRepo,
		feedbackRepo:   feedbackRepo,
		reportRepo:     reportRepo,
	}
}

// HistoryListInput 是训练历史列表查询输入。
type HistoryListInput struct {
	UserID   int
	Page     int
	PageSize int
}

// HistoryListResult 是训练历史列表查询输出。
type HistoryListResult struct {
	Items    []HistoryItem
	Page     int
	PageSize int
	Total    int
}

// HistoryItem 是历史列表中的单条训练摘要。
type HistoryItem struct {
	Session         model.Session
	Scenario        model.Scenario
	TotalScore      *int
	ReportGenerated bool
}

// ListSessions 查询训练历史列表。
func (s *HistoryService) ListSessions(input HistoryListInput) (HistoryListResult, error) {
	page, pageSize, err := normalizeHistoryPagination(input.Page, input.PageSize)
	if err != nil {
		return HistoryListResult{}, err
	}
	if input.UserID < 0 {
		return HistoryListResult{}, ErrInvalidHistoryRequest
	}

	list, err := s.sessionRepo.ListSessions(model.SessionListQuery{
		UserID:   input.UserID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return HistoryListResult{}, err
	}

	items := make([]HistoryItem, 0, len(list.Sessions))
	for _, session := range list.Sessions {
		scenario, err := s.scenarioReader.GetScenario(session.ScenarioID)
		if err != nil {
			return HistoryListResult{}, err
		}
		totalScore, err := s.currentTotalScore(session.ID)
		if err != nil {
			return HistoryListResult{}, err
		}
		reportGenerated, err := s.hasReport(session.ID)
		if err != nil {
			return HistoryListResult{}, err
		}
		items = append(items, HistoryItem{
			Session:         session,
			Scenario:        scenario,
			TotalScore:      totalScore,
			ReportGenerated: reportGenerated,
		})
	}

	return HistoryListResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    list.Total,
	}, nil
}

// normalizeHistoryPagination 归一化历史列表分页参数。
func normalizeHistoryPagination(page int, pageSize int) (int, int, error) {
	if page == 0 {
		page = defaultHistoryPage
	}
	if pageSize == 0 {
		pageSize = defaultHistoryPageSize
	}
	if page < 0 || pageSize < 0 {
		return 0, 0, ErrInvalidHistoryRequest
	}
	if pageSize > maxHistoryPageSize {
		pageSize = maxHistoryPageSize
	}

	return page, pageSize, nil
}

// currentTotalScore 读取历史摘要中的当前总分。
func (s *HistoryService) currentTotalScore(sessionID int) (*int, error) {
	if s.feedbackRepo == nil {
		return nil, nil
	}
	score, err := s.feedbackRepo.FindCurrentScoreBySessionID(sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrScoreNotFound) || errors.Is(err, ErrScoreNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &score.TotalScore, nil
}

// hasReport 判断指定 Session 是否已有报告。
func (s *HistoryService) hasReport(sessionID int) (bool, error) {
	if s.reportRepo == nil {
		return false, nil
	}
	_, err := s.reportRepo.FindBySessionID(sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrReportNotFound) || errors.Is(err, ErrReportNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
