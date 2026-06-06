package agent

import (
	"math"

	"speakmate/internal/model"
)

// MockScoringAgent 基于纠错结果返回稳定的本地 Mock 评分。
type MockScoringAgent struct{}

// NewMockScoringAgent 创建 Mock 评分 Agent。
func NewMockScoringAgent() *MockScoringAgent {
	return &MockScoringAgent{}
}

// Score 根据是否存在纠错问题返回固定分项评分。
func (a *MockScoringAgent) Score(input ScoringInput) (ScoringOutput, error) {
	messageID := input.Correction.MessageID
	if messageID == 0 {
		messageID = input.UserMessage.ID
	}

	sessionID := input.Correction.SessionID
	if sessionID == 0 {
		sessionID = input.UserMessage.SessionID
	}
	if sessionID == 0 {
		sessionID = input.Session.ID
	}

	result := model.ScoreResult{
		MessageID: messageID,
		SessionID: sessionID,
	}
	if len(input.Correction.Errors) > 0 {
		result.Fluency = 75
		result.Grammar = 72
		result.Expression = 80
		result.Vocabulary = 76
		result.Completion = 85
		result.Comment = "用户能够表达核心意思，但存在时态和动词形式错误。"
	} else {
		result.Fluency = 85
		result.Grammar = 88
		result.Expression = 84
		result.Vocabulary = 82
		result.Completion = 86
		result.Comment = "用户表达清晰，当前轮次没有明显语法或用词问题。"
	}
	result.TotalScore = weightedTotalScore(result)

	return ScoringOutput{
		Result: result,
		Raw:    nil,
	}, nil
}

func weightedTotalScore(result model.ScoreResult) int {
	return int(math.Round(
		0.25*float64(result.Fluency) +
			0.25*float64(result.Grammar) +
			0.20*float64(result.Expression) +
			0.15*float64(result.Vocabulary) +
			0.15*float64(result.Completion),
	))
}
