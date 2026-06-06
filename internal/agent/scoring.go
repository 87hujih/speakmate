package agent

import "speakmate/internal/model"

// ScoringAgent 定义用户英文表达评分能力。
type ScoringAgent interface {
	Score(input ScoringInput) (ScoringOutput, error)
}

// ScoringInput 是生成单条用户消息评分所需的上下文。
type ScoringInput struct {
	Scenario    model.Scenario
	Session     model.Session
	History     []model.Message
	UserMessage model.Message
	Correction  model.CorrectionResult
}

// ScoringOutput 是 Scoring Agent 的结构化输出。
type ScoringOutput struct {
	Result model.ScoreResult
	Raw    any
}
