package model

import "time"

// ReportScenario 是报告中使用的场景摘要。
type ReportScenario struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Difficulty string `json:"difficulty"`
}

// Report 表示一次训练完成后的结构化课后报告。
type Report struct {
	SessionID         int            `json:"session_id"`
	Scenario          ReportScenario `json:"scenario"`
	DurationSeconds   int            `json:"duration_seconds"`
	TurnCount         int            `json:"turn_count"`
	TotalScore        int            `json:"total_score"`
	Scores            ScoreResult    `json:"scores"`
	Summary           string         `json:"summary"`
	MajorProblems     []string       `json:"major_problems"`
	FrequentErrors    []string       `json:"frequent_errors"`
	BetterExpressions []string       `json:"better_expressions"`
	NextPracticePlan  []string       `json:"next_practice_plan"`
	CreatedAt         time.Time      `json:"created_at"`
}
