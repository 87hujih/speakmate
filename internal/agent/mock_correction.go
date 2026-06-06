package agent

import (
	"strings"

	"speakmate/internal/model"
)

// MockCorrectionAgent 基于本地规则返回稳定的纠错结果。
type MockCorrectionAgent struct{}

// NewMockCorrectionAgent 创建 Mock 纠错 Agent。
func NewMockCorrectionAgent() *MockCorrectionAgent {
	return &MockCorrectionAgent{}
}

// Correct 使用少量固定规则生成可预测的纠错输出。
func (a *MockCorrectionAgent) Correct(input CorrectionInput) (CorrectionOutput, error) {
	original := strings.TrimSpace(input.UserMessage.Content)
	corrected := original
	errors := make([]model.CorrectionError, 0)

	if strings.Contains(corrected, "am study") {
		corrected = strings.Replace(corrected, "am study", "am studying", 1)
		errors = append(errors, model.CorrectionError{
			Type:        model.CorrectionErrorTypeGrammar,
			Span:        "am study",
			Suggestion:  "am studying",
			Explanation: "be 动词后应接现在分词。",
		})
	}
	if strings.Contains(corrected, "have did") {
		corrected = strings.Replace(corrected, "have did", "have done", 1)
		errors = append(errors, model.CorrectionError{
			Type:        model.CorrectionErrorTypeGrammar,
			Span:        "have did",
			Suggestion:  "have done",
			Explanation: "现在完成时中 have 后应接过去分词 done。",
		})
	}
	if len(errors) > 0 && strings.Contains(corrected, "computer science and") {
		corrected = strings.Replace(corrected, "computer science and", "computer science, and", 1)
	}

	sessionID := input.UserMessage.SessionID
	if sessionID == 0 {
		sessionID = input.Session.ID
	}

	return CorrectionOutput{
		Result: model.CorrectionResult{
			MessageID:         input.UserMessage.ID,
			SessionID:         sessionID,
			OriginalText:      original,
			CorrectedText:     corrected,
			Errors:            errors,
			BetterExpressions: betterExpressionsForScenario(input.Scenario.Code),
		},
		Raw: nil,
	}, nil
}

func betterExpressionsForScenario(code string) []string {
	switch code {
	case "interview":
		return []string{
			"I major in computer science.",
			"I worked on a robotics project.",
		}
	case "restaurant":
		return []string{
			"Could you recommend a house special?",
			"I would like something light and not too spicy.",
		}
	case "meeting":
		return []string{
			"The main blocker is the API integration timeline.",
			"I recommend this option because it reduces risk.",
		}
	default:
		return []string{
			"Could you add one specific detail to make the idea clearer?",
		}
	}
}
