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
	majorProblems := majorProblemsFromFeedback(input.Score, input.Corrections, input.HistoryMessages())
	nextPracticePlan := nextPracticePlanForSummary(input.Scenario.Code, input.Corrections, input.HistoryMessages())

	scenarioName := strings.TrimSpace(input.Scenario.Name)
	if scenarioName == "" {
		scenarioName = "本次"
	}
	userEvidence := firstUserUtterance(input.HistoryMessages())
	if userEvidence == "" {
		userEvidence = "本次对话"
	}
	summary := fmt.Sprintf(
		"%s训练完成 %d 轮，当前综合评分 %d。用户原话“%s”是本次分析的主要证据，%s",
		scenarioName,
		input.Session.TurnCount,
		input.Score.TotalScore,
		truncateReportText(userEvidence, 120),
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
			explanation := strings.TrimSpace(correctionError.Explanation)
			if explanation != "" {
				item += " | 原因：" + explanation
			}
			original := strings.TrimSpace(correction.OriginalText)
			if original != "" {
				item += " | 证据：“" + truncateReportText(original, 90) + "”"
			}
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
		original := strings.TrimSpace(correction.OriginalText)
		for _, expression := range correction.BetterExpressions {
			expression = strings.TrimSpace(expression)
			if expression == "" {
				continue
			}
			if original != "" {
				expression = truncateReportText(original, 90) + " -> " + expression
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

func majorProblemsFromFeedback(score model.ScoreResult, corrections []model.CorrectionResult, history []model.Message) []string {
	problems := make([]string, 0)
	evidence := firstCorrectionEvidence(corrections)
	if evidence == "" {
		evidence = firstUserUtterance(history)
	}
	if score.Grammar > 0 && score.Grammar < 80 {
		problems = append(problems, withEvidence("语法准确度需要加强，优先检查动词形式和时态。", evidence))
	}
	if score.Expression > 0 && score.Expression < 80 {
		problems = append(problems, withEvidence("表达自然度还有提升空间，可以多使用场景化短语。", evidence))
	}
	if score.Completion > 0 && score.Completion < 80 {
		problems = append(problems, withEvidence("场景任务完成度不足，回答中需要补充更多具体信息。", evidence))
	}
	if len(problems) == 0 && hasCorrectionErrors(corrections) {
		problems = append(problems, withEvidence("表达中存在可修正的小错误，建议复盘纠错列表。", evidence))
	}
	if len(problems) == 0 {
		problems = append(problems, withEvidence("本次表达整体清晰，下一步可以提高回答细节和自然度。", evidence))
	}

	return problems
}

func nextPracticePlanForSummary(code string, corrections []model.CorrectionResult, history []model.Message) []string {
	correctionTarget := firstCorrectionTarget(corrections)
	switch code {
	case "interview":
		return appendCorrectionPractice([]string{
			"任务：用 STAR 结构重写一次项目经历回答 | 验收：包含 Situation、Task、Action、Result 各 1 句。",
			"任务：准备 3 个关于个人贡献和技术难点的英文追问回答 | 验收：每个回答至少包含一个具体动作和一个结果。",
		}, correctionTarget)
	case "restaurant":
		return appendCorrectionPractice([]string{
			"任务：练习 5 组点餐偏好表达，例如口味、过敏和推荐请求 | 验收：每组都使用 Could you 或 I would like。",
			"任务：复述一次完整点餐流程，注意礼貌请求句型 | 验收：从入座到结账至少 6 句完整英文。",
		}, correctionTarget)
	case "meeting":
		return appendCorrectionPractice([]string{
			"任务：准备 3 句表达进度、风险和建议的会议句型 | 验收：每句包含一个业务名词和一个下一步动作。",
			"任务：练习用一句话总结自己的观点和下一步行动 | 验收：控制在 20 个英文词以内。",
		}, correctionTarget)
	default:
		if correctionTarget != "" {
			return []string{correctionTarget}
		}

		evidence := firstUserUtterance(history)
		if evidence != "" {
			return []string{"任务：围绕“" + truncateReportText(evidence, 60) + "”补充 2 个具体细节 | 验收：下一轮回答至少包含一个数字、一个动作和一个结果。"}
		}

		return []string{"任务：下一轮练习中至少补充 2 个具体细节，让回答更完整 | 验收：回答至少 4 句，并覆盖场景目标。"}
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

func firstUserUtterance(history []model.Message) string {
	for _, message := range history {
		if message.Role != model.MessageRoleUser {
			continue
		}
		content := strings.TrimSpace(message.Content)
		if content != "" {
			return content
		}
	}

	return ""
}

func firstCorrectionEvidence(corrections []model.CorrectionResult) string {
	for _, correction := range corrections {
		original := strings.TrimSpace(correction.OriginalText)
		for _, correctionError := range correction.Errors {
			span := strings.TrimSpace(correctionError.Span)
			suggestion := strings.TrimSpace(correctionError.Suggestion)
			if span == "" || suggestion == "" {
				continue
			}
			evidence := span + " -> " + suggestion
			if original != "" {
				evidence += "，原句：“" + truncateReportText(original, 80) + "”"
			}

			return evidence
		}
	}

	return ""
}

func firstCorrectionTarget(corrections []model.CorrectionResult) string {
	pairs := make([]string, 0)
	suggestions := make([]string, 0)
	seen := make(map[string]struct{})
	for _, correction := range corrections {
		for _, correctionError := range correction.Errors {
			span := strings.TrimSpace(correctionError.Span)
			suggestion := strings.TrimSpace(correctionError.Suggestion)
			if span == "" || suggestion == "" {
				continue
			}
			key := span + "->" + suggestion
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			pairs = append(pairs, fmt.Sprintf("“%s” 改写为 “%s”", span, suggestion))
			suggestions = append(suggestions, "“"+suggestion+"”")
		}
	}
	if len(pairs) == 0 {
		return ""
	}

	return fmt.Sprintf(
		"任务：把 %s 并各造 2 个新句子 | 验收：新句子必须正确使用 %s。",
		strings.Join(pairs, "、"),
		strings.Join(suggestions, "、"),
	)
}

func appendCorrectionPractice(items []string, correctionTarget string) []string {
	if correctionTarget == "" {
		return items
	}

	return append([]string{correctionTarget}, items...)
}

func withEvidence(problem string, evidence string) string {
	evidence = strings.TrimSpace(evidence)
	if evidence == "" {
		return problem
	}

	return problem + " 证据：" + evidence
}

func truncateReportText(text string, maxLen int) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if maxLen <= 0 || len(text) <= maxLen {
		return text
	}

	return text[:maxLen] + "..."
}
