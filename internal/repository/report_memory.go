package repository

import (
	"errors"
	"sync"

	"speakmate/internal/model"
)

var (
	// ErrReportNotFound 表示内存仓库中没有找到对应报告。
	ErrReportNotFound = errors.New("report not found")
)

// MemoryReportRepository 使用内存 map 保存训练报告。
type MemoryReportRepository struct {
	mu                 sync.RWMutex
	reportsBySessionID map[int]model.Report
}

// NewMemoryReportRepository 创建空的内存 Report 仓库。
func NewMemoryReportRepository() *MemoryReportRepository {
	return &MemoryReportRepository{
		reportsBySessionID: make(map[int]model.Report),
	}
}

// Save 按 session_id 保存或覆盖报告。
func (r *MemoryReportRepository) Save(report model.Report) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reportsBySessionID[report.SessionID] = cloneReport(report)

	return nil
}

// FindBySessionID 按 session_id 查询报告。
func (r *MemoryReportRepository) FindBySessionID(sessionID int) (model.Report, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report, ok := r.reportsBySessionID[sessionID]
	if !ok {
		return model.Report{}, ErrReportNotFound
	}

	return cloneReport(report), nil
}

func cloneReport(report model.Report) model.Report {
	if report.MajorProblems != nil {
		report.MajorProblems = append([]string(nil), report.MajorProblems...)
	}
	if report.FrequentErrors != nil {
		report.FrequentErrors = append([]string(nil), report.FrequentErrors...)
	}
	if report.BetterExpressions != nil {
		report.BetterExpressions = append([]string(nil), report.BetterExpressions...)
	}
	if report.NextPracticePlan != nil {
		report.NextPracticePlan = append([]string(nil), report.NextPracticePlan...)
	}

	return report
}
