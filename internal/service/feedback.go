package service

import (
	"errors"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

// 服务层复用的哨兵错误。
var (
	// ErrInvalidFeedbackRequest 表示反馈查询参数非法。
	ErrInvalidFeedbackRequest = errors.New("反馈请求无效")
	// ErrCorrectionNotFound 表示业务层没有找到对应纠错结果。
	ErrCorrectionNotFound = errors.New("未找到纠错结果")
	// ErrScoreNotFound 表示业务层没有找到对应评分结果。
	ErrScoreNotFound = errors.New("未找到评分结果")
)

// FeedbackRepository 定义 Feedback 服务依赖的数据访问能力。
type FeedbackRepository interface {
	SaveCorrection(correction model.CorrectionResult) error
	SaveScore(score model.ScoreResult) error
	FindCorrectionByMessageID(messageID int) (model.CorrectionResult, error)
	ListCorrectionsBySessionID(sessionID int) ([]model.CorrectionResult, error)
	FindCurrentScoreBySessionID(sessionID int) (model.ScoreResult, error)
}

// FeedbackService 封装纠错和评分查询业务流程。
type FeedbackService struct {
	repo FeedbackRepository
}

// NewFeedbackService 创建 Feedback 服务实例。
func NewFeedbackService(repo FeedbackRepository) *FeedbackService {
	return &FeedbackService{
		repo: repo,
	}
}

// GetMessageCorrection 按 message_id 查询单条消息纠错。
func (s *FeedbackService) GetMessageCorrection(messageID int) (model.CorrectionResult, error) {
	if messageID <= 0 {
		return model.CorrectionResult{}, ErrInvalidFeedbackRequest
	}

	correction, err := s.repo.FindCorrectionByMessageID(messageID)
	if err == nil {
		return correction, nil
	}
	if errors.Is(err, repository.ErrCorrectionNotFound) {
		return model.CorrectionResult{}, ErrCorrectionNotFound
	}

	return model.CorrectionResult{}, err
}

// ListSessionCorrections 按 session_id 查询整场训练的全部纠错。
func (s *FeedbackService) ListSessionCorrections(sessionID int) ([]model.CorrectionResult, error) {
	if sessionID <= 0 {
		return nil, ErrInvalidFeedbackRequest
	}

	corrections, err := s.repo.ListCorrectionsBySessionID(sessionID)
	if err == nil {
		return corrections, nil
	}
	if errors.Is(err, repository.ErrCorrectionNotFound) {
		return nil, ErrCorrectionNotFound
	}

	return nil, err
}

// GetSessionCurrentScore 按 session_id 查询当前评分。
func (s *FeedbackService) GetSessionCurrentScore(sessionID int) (model.ScoreResult, error) {
	if sessionID <= 0 {
		return model.ScoreResult{}, ErrInvalidFeedbackRequest
	}

	score, err := s.repo.FindCurrentScoreBySessionID(sessionID)
	if err == nil {
		return score, nil
	}
	if errors.Is(err, repository.ErrScoreNotFound) {
		return model.ScoreResult{}, ErrScoreNotFound
	}

	return model.ScoreResult{}, err
}
