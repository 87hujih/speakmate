package repository

import (
	"errors"
	"sync"

	"speakmate/internal/model"
)

var (
	// ErrCorrectionNotFound 表示内存仓库中没有找到对应纠错结果。
	ErrCorrectionNotFound = errors.New("correction not found")
	// ErrScoreNotFound 表示内存仓库中没有找到对应评分结果。
	ErrScoreNotFound = errors.New("score not found")
)

// MemoryFeedbackRepository 使用内存 map 保存消息纠错和 Session 当前评分。
type MemoryFeedbackRepository struct {
	mu                      sync.RWMutex
	correctionsByMessageID  map[int]model.CorrectionResult
	correctionsBySessionID  map[int][]model.CorrectionResult
	scoresByMessageID       map[int]model.ScoreResult
	currentScoreBySessionID map[int]model.ScoreResult
}

// NewMemoryFeedbackRepository 创建空的内存 Feedback 仓库。
func NewMemoryFeedbackRepository() *MemoryFeedbackRepository {
	return &MemoryFeedbackRepository{
		correctionsByMessageID:  make(map[int]model.CorrectionResult),
		correctionsBySessionID:  make(map[int][]model.CorrectionResult),
		scoresByMessageID:       make(map[int]model.ScoreResult),
		currentScoreBySessionID: make(map[int]model.ScoreResult),
	}
}

// SaveCorrection 保存或覆盖单条消息的纠错结果。
func (r *MemoryFeedbackRepository) SaveCorrection(correction model.CorrectionResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if previous, ok := r.correctionsByMessageID[correction.MessageID]; ok && previous.SessionID != correction.SessionID {
		r.removeCorrectionFromSession(previous.SessionID, correction.MessageID)
	}

	cloned := cloneCorrectionResult(correction)
	r.correctionsByMessageID[correction.MessageID] = cloned
	r.upsertSessionCorrection(correction.SessionID, cloned)

	return nil
}

// SaveScore 保存单条消息评分，并把它设为该 Session 当前评分。
func (r *MemoryFeedbackRepository) SaveScore(score model.ScoreResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.scoresByMessageID[score.MessageID] = score
	r.currentScoreBySessionID[score.SessionID] = score

	return nil
}

// FindCorrectionByMessageID 按 message_id 查询单条纠错结果。
func (r *MemoryFeedbackRepository) FindCorrectionByMessageID(messageID int) (model.CorrectionResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	correction, ok := r.correctionsByMessageID[messageID]
	if !ok {
		return model.CorrectionResult{}, ErrCorrectionNotFound
	}

	return cloneCorrectionResult(correction), nil
}

// ListCorrectionsBySessionID 按 session_id 查询全部纠错结果。
func (r *MemoryFeedbackRepository) ListCorrectionsBySessionID(sessionID int) ([]model.CorrectionResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	corrections, ok := r.correctionsBySessionID[sessionID]
	if !ok || len(corrections) == 0 {
		return nil, ErrCorrectionNotFound
	}

	return cloneCorrectionResults(corrections), nil
}

// FindCurrentScoreBySessionID 按 session_id 查询当前评分。
func (r *MemoryFeedbackRepository) FindCurrentScoreBySessionID(sessionID int) (model.ScoreResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	score, ok := r.currentScoreBySessionID[sessionID]
	if !ok {
		return model.ScoreResult{}, ErrScoreNotFound
	}

	return score, nil
}

func (r *MemoryFeedbackRepository) upsertSessionCorrection(sessionID int, correction model.CorrectionResult) {
	corrections := r.correctionsBySessionID[sessionID]
	for i, existing := range corrections {
		if existing.MessageID == correction.MessageID {
			corrections[i] = correction
			r.correctionsBySessionID[sessionID] = corrections
			return
		}
	}

	r.correctionsBySessionID[sessionID] = append(corrections, correction)
}

func (r *MemoryFeedbackRepository) removeCorrectionFromSession(sessionID int, messageID int) {
	corrections := r.correctionsBySessionID[sessionID]
	for i, existing := range corrections {
		if existing.MessageID == messageID {
			corrections = append(corrections[:i], corrections[i+1:]...)
			break
		}
	}
	if len(corrections) == 0 {
		delete(r.correctionsBySessionID, sessionID)
		return
	}

	r.correctionsBySessionID[sessionID] = corrections
}

func cloneCorrectionResults(corrections []model.CorrectionResult) []model.CorrectionResult {
	cloned := make([]model.CorrectionResult, len(corrections))
	for i, correction := range corrections {
		cloned[i] = cloneCorrectionResult(correction)
	}

	return cloned
}

func cloneCorrectionResult(correction model.CorrectionResult) model.CorrectionResult {
	if correction.Errors != nil {
		correction.Errors = append([]model.CorrectionError(nil), correction.Errors...)
	}
	if correction.BetterExpressions != nil {
		correction.BetterExpressions = append([]string(nil), correction.BetterExpressions...)
	}

	return correction
}
