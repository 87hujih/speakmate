package repository_test

import (
	"errors"
	"testing"
	"time"

	"speakmate/internal/model"
	"speakmate/internal/repository"
)

func TestMemoryReportRepositorySavesAndFindsBySessionID(t *testing.T) {
	repo := repository.NewMemoryReportRepository()
	report := sampleReport(7)

	if err := repo.Save(report); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	found, err := repo.FindBySessionID(7)
	if err != nil {
		t.Fatalf("FindBySessionID returned error: %v", err)
	}

	if found.SessionID != 7 {
		t.Fatalf("session_id = %d, want 7", found.SessionID)
	}
	if found.Scenario.Code != "interview" {
		t.Fatalf("scenario code = %q, want interview", found.Scenario.Code)
	}
	if found.TotalScore != 77 {
		t.Fatalf("total_score = %d, want 77", found.TotalScore)
	}
	if found.Summary != "用户能够完成面试核心表达，但语法准确度需要加强。" {
		t.Fatalf("summary = %q, want saved summary", found.Summary)
	}
	if len(found.FrequentErrors) != 1 || found.FrequentErrors[0] != "am study -> am studying" {
		t.Fatalf("frequent_errors = %#v, want saved errors", found.FrequentErrors)
	}
}

func TestMemoryReportRepositoryOverwritesSameSessionReport(t *testing.T) {
	repo := repository.NewMemoryReportRepository()
	first := sampleReport(7)
	second := sampleReport(7)
	second.TotalScore = 88
	second.Summary = "第二次生成覆盖旧报告。"

	if err := repo.Save(first); err != nil {
		t.Fatalf("Save first returned error: %v", err)
	}
	if err := repo.Save(second); err != nil {
		t.Fatalf("Save second returned error: %v", err)
	}

	found, err := repo.FindBySessionID(7)
	if err != nil {
		t.Fatalf("FindBySessionID returned error: %v", err)
	}

	if found.TotalScore != 88 {
		t.Fatalf("total_score = %d, want overwritten score 88", found.TotalScore)
	}
	if found.Summary != "第二次生成覆盖旧报告。" {
		t.Fatalf("summary = %q, want overwritten summary", found.Summary)
	}
}

func TestMemoryReportRepositoryReturnsNotFound(t *testing.T) {
	repo := repository.NewMemoryReportRepository()

	_, err := repo.FindBySessionID(404)

	if !errors.Is(err, repository.ErrReportNotFound) {
		t.Fatalf("FindBySessionID error = %v, want ErrReportNotFound", err)
	}
}

func TestMemoryReportRepositoryReturnsCopies(t *testing.T) {
	repo := repository.NewMemoryReportRepository()
	report := sampleReport(7)

	if err := repo.Save(report); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	report.FrequentErrors[0] = "mutated before query"
	report.BetterExpressions[0] = "mutated before query"
	report.NextPracticePlan[0] = "mutated before query"

	found, err := repo.FindBySessionID(7)
	if err != nil {
		t.Fatalf("FindBySessionID returned error: %v", err)
	}
	found.FrequentErrors[0] = "mutated after query"
	found.BetterExpressions[0] = "mutated after query"
	found.NextPracticePlan[0] = "mutated after query"

	again, err := repo.FindBySessionID(7)
	if err != nil {
		t.Fatalf("FindBySessionID returned error: %v", err)
	}

	if again.FrequentErrors[0] != "am study -> am studying" {
		t.Fatalf("stored frequent error = %q, want original", again.FrequentErrors[0])
	}
	if again.BetterExpressions[0] != "I am studying computer science." {
		t.Fatalf("stored better expression = %q, want original", again.BetterExpressions[0])
	}
	if again.NextPracticePlan[0] != "用现在进行时重写 5 个自我介绍句子。" {
		t.Fatalf("stored next practice plan = %q, want original", again.NextPracticePlan[0])
	}
}

func sampleReport(sessionID int) model.Report {
	return model.Report{
		SessionID: sessionID,
		Scenario: model.ReportScenario{
			ID:         1,
			Code:       "interview",
			Name:       "英语面试",
			Difficulty: "medium",
		},
		DurationSeconds: 180,
		TurnCount:       2,
		TotalScore:      77,
		Scores: model.ScoreResult{
			SessionID:  sessionID,
			Grammar:    72,
			Expression: 80,
			TotalScore: 77,
			Comment:    "用户能够表达核心意思，但存在动词形式错误。",
		},
		Summary:           "用户能够完成面试核心表达，但语法准确度需要加强。",
		MajorProblems:     []string{"动词形式不准确"},
		FrequentErrors:    []string{"am study -> am studying"},
		BetterExpressions: []string{"I am studying computer science."},
		NextPracticePlan:  []string{"用现在进行时重写 5 个自我介绍句子。"},
		CreatedAt:         time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC),
	}
}
