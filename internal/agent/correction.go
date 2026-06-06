package agent

import "speakmate/internal/model"

// CorrectionAgent 定义用户英文表达纠错能力。
type CorrectionAgent interface {
	Correct(input CorrectionInput) (CorrectionOutput, error)
}

// CorrectionInput 是生成单条用户消息纠错结果所需的上下文。
type CorrectionInput struct {
	Scenario    model.Scenario
	Session     model.Session
	History     []model.Message
	UserMessage model.Message
}

// CorrectionOutput 是 Correction Agent 的结构化输出。
type CorrectionOutput struct {
	Result model.CorrectionResult
	Raw    any
}
