package agent

import "speakmate/internal/model"

// SummaryAgent 定义训练结束后的课后报告总结能力。
type SummaryAgent interface {
	Summarize(input SummaryInput) (SummaryOutput, error)
}

// SummaryInput 是生成课后报告总结所需的完整训练上下文。
type SummaryInput struct {
	Scenario    model.Scenario
	Session     model.Session
	Messages    []model.Message
	Corrections []model.CorrectionResult
	Score       model.ScoreResult
}

// SummaryOutput 是 Summary Agent 生成的结构化报告内容。
type SummaryOutput struct {
	Summary           string
	MajorProblems     []string
	FrequentErrors    []string
	BetterExpressions []string
	NextPracticePlan  []string
	Raw               any
}

func (input SummaryInput) HistoryMessages() []model.Message {
	if input.Messages != nil {
		return input.Messages
	}

	return input.Session.Messages
}
