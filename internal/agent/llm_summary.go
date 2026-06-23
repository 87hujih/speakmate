package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"speakmate/internal/infra/llm"
)

// LLMSummaryAgent 使用 LLM 生成课后报告摘要。
type LLMSummaryAgent struct {
	client   llm.Client
	fallback SummaryAgent
}

// LLMSummaryOption 用于配置 LLMSummaryAgent。
type LLMSummaryOption func(*LLMSummaryAgent)

// NewLLMSummaryAgent 创建并返回对应组件实例。
func NewLLMSummaryAgent(client llm.Client, opts ...LLMSummaryOption) *LLMSummaryAgent {
	agent := &LLMSummaryAgent{
		client: client,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// WithSummaryFallbackAgent 返回用于覆盖默认行为的配置选项。
func WithSummaryFallbackAgent(fallback SummaryAgent) LLMSummaryOption {
	return func(agent *LLMSummaryAgent) {
		agent.fallback = fallback
	}
}

// Summarize 调用 LLM 生成结构化课后报告摘要，并在失败时按配置降级。
func (a *LLMSummaryAgent) Summarize(input SummaryInput) (SummaryOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(input, errors.New("LLM 客户端不能为空"))
	}

	response, err := a.client.CreateChatCompletion(context.Background(), llm.ChatRequest{
		Messages: toLLMMessages(BuildSummaryPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(input, fmt.Errorf("创建聊天补全失败：%w", err))
	}

	result, err := parseSummaryJSON(response.Content)
	if err != nil {
		return a.fallbackOrError(input, err)
	}
	result.Raw = response.Raw

	return result, nil
}

// fallbackOrError 在配置 fallback 时降级处理，否则返回原始错误。
func (a *LLMSummaryAgent) fallbackOrError(input SummaryInput, err error) (SummaryOutput, error) {
	if a.fallback != nil {
		return a.fallback.Summarize(input)
	}

	return SummaryOutput{}, err
}

// summaryPayload 是 LLM 报告摘要 JSON 的内部解析结构。
type summaryPayload struct {
	Summary           *string   `json:"summary"`
	MajorProblems     *[]string `json:"major_problems"`
	FrequentErrors    *[]string `json:"frequent_errors"`
	BetterExpressions *[]string `json:"better_expressions"`
	NextPracticePlan  *[]string `json:"next_practice_plan"`
}

// parseSummaryJSON 严格解析 LLM 返回的报告摘要 JSON。
func parseSummaryJSON(content string) (SummaryOutput, error) {
	var payload summaryPayload
	if err := decodeStrictJSONObject(content, &payload); err != nil {
		return SummaryOutput{}, fmt.Errorf("解析报告摘要 JSON 失败：%w", err)
	}

	summary, err := requiredTrimmedString(payload.Summary, "summary summary")
	if err != nil {
		return SummaryOutput{}, err
	}
	majorProblems, err := requiredStringList(payload.MajorProblems, "summary major_problems")
	if err != nil {
		return SummaryOutput{}, err
	}
	frequentErrors, err := requiredStringList(payload.FrequentErrors, "summary frequent_errors")
	if err != nil {
		return SummaryOutput{}, err
	}
	betterExpressions, err := requiredStringList(payload.BetterExpressions, "summary better_expressions")
	if err != nil {
		return SummaryOutput{}, err
	}
	nextPracticePlan, err := requiredStringList(payload.NextPracticePlan, "summary next_practice_plan")
	if err != nil {
		return SummaryOutput{}, err
	}

	return SummaryOutput{
		Summary:           summary,
		MajorProblems:     majorProblems,
		FrequentErrors:    frequentErrors,
		BetterExpressions: betterExpressions,
		NextPracticePlan:  nextPracticePlan,
	}, nil
}

// requiredStringList 校验报告摘要中的必填字符串数组。
func requiredStringList(value *[]string, field string) ([]string, error) {
	if value == nil {
		return nil, fmt.Errorf("%s 不能为空", field)
	}

	result := make([]string, 0, len(*value))
	for i, item := range *value {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("%s[%d] 不能为空", field, i)
		}
		result = append(result, item)
	}

	return result, nil
}
