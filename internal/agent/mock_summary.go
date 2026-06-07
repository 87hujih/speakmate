package agent

import (
	"fmt"
	"strings"

	"speakmate/internal/model"
)

// MockSummaryAgent 基于纠错和评分结果返回稳定的课后报告摘要。
type MockSummaryAgent struct{}

// NewMockSummaryAgent 创建 Mock Summary Agent。
func NewMockSummaryAgent() *MockSummaryAgent {
	return &MockSummaryAgent{}
}

// Summarize 从本地反馈数据中归纳报告内容。
func (a *MockSummaryAgent) Summarize(input SummaryInput) (SummaryOutput, error) {
	frequentErrors := frequentErrorsFromCorrections(input.Corrections)
	betterExpressions := betterExpressionsFromCorrections(input.Corrections)
	majorProblems := majorProblemsFromFeedback(input.Score, input.Corrections)
	nextPracticePlan := nextPracticePlanForSummary(input.Scenario.Code, len(input.Corrections) > 0)

	scenarioName := strings.TrimSpace(input.Scenario.Name)
	if scenarioName == "" {
		scenarioName = "本次"
	}
	summary := fmt.Sprintf(
		"%s训练完成 %d 轮，当前综合评分 %d。%s",
		scenarioName,
		input.Session.TurnCount,
		input.Score.TotalScore,
		scoreCommentOrFallback(input.Score.Comment),
	)

	return SummaryOutput{
		Summary:           summary,
		MajorProblems:     majorProblems,
		FrequentErrors:    frequentErrors,
		BetterExpressions: betterExpressions,
		NextPracticePlan:  nextPracticePlan,
		Raw:               nil,
	}, nil
}

func frequentErrorsFromCorrections(corrections []model.CorrectionResult) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, correction := range corrections {
		for _, correctionError := range correction.Errors {
			span := strings.TrimSpace(correctionError.Span)
			suggestion := strings.TrimSpace(correctionError.Suggestion)
			if span == "" || suggestion == "" {
				continue
			}
			item := span + " -> " + suggestion
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return []string{"暂未发现高频错误，下一轮继续保持句子完整度和场景相关性。"}
	}

	return result
}

func betterExpressionsFromCorrections(corrections []model.CorrectionResult) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, correction := range corrections {
		for _, expression := range correction.BetterExpressions {
			expression = strings.TrimSpace(expression)
			if expression == "" {
				continue
			}
			if _, ok := seen[expression]; ok {
				continue
			}
			seen[expression] = struct{}{}
			result = append(result, expression)
		}
	}
	if len(result) == 0 {
		return []string{"Try adding one concrete example to make your answer more specific."}
	}

	return result
}

func majorProblemsFromFeedback(score model.ScoreResult, corrections []model.CorrectionResult) []string {
	problems := make([]string, 0)
	if score.Grammar > 0 && score.Grammar < 80 {
		problems = append(problems, "语法准确度需要加强，优先检查动词形式和时态。")
	}
	if score.Expression > 0 && score.Expression < 80 {
		problems = append(problems, "表达自然度还有提升空间，可以多使用场景化短语。")
	}
	if score.Completion > 0 && score.Completion < 80 {
		problems = append(problems, "场景任务完成度不足，回答中需要补充更多具体信息。")
	}
	if len(problems) == 0 && hasCorrectionErrors(corrections) {
		problems = append(problems, "表达中存在可修正的小错误，建议复盘纠错列表。")
	}
	if len(problems) == 0 {
		problems = append(problems, "本次表达整体清晰，下一步可以提高回答细节和自然度。")
	}

	return problems
}

func nextPracticePlanForSummary(code string, hasCorrections bool) []string {
	switch code {
	case "interview":
		return []string{
			"用 STAR 结构重写一次项目经历回答。",
			"准备 3 个关于个人贡献和技术难点的英文追问回答。",
		}
	case "restaurant":
		return []string{
			"练习 5 组点餐偏好表达，例如口味、过敏和推荐请求。",
			"复述一次完整点餐流程，注意礼貌请求句型。",
		}
	case "meeting":
		return []string{
			"准备 3 句表达进度、风险和建议的会议句型。",
			"练习用一句话总结自己的观点和下一步行动。",
		}
	default:
		if hasCorrections {
			return []string{"重读本次纠错句子，并用正确表达各造 2 个新句子。"}
		}

		return []string{"下一轮练习中至少补充 2 个具体细节，让回答更完整。"}
	}
}

func scoreCommentOrFallback(comment string) string {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return "建议继续围绕场景目标补充更具体的表达。"
	}

	return comment
}

func hasCorrectionErrors(corrections []model.CorrectionResult) bool {
	for _, correction := range corrections {
		if len(correction.Errors) > 0 {
			return true
		}
	}

	return false
}
