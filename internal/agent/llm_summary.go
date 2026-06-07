package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"speakmate/internal/infra/llm"
)

type LLMSummaryAgent struct {
	client   llm.Client
	fallback SummaryAgent
}

type LLMSummaryOption func(*LLMSummaryAgent)

func NewLLMSummaryAgent(client llm.Client, opts ...LLMSummaryOption) *LLMSummaryAgent {
	agent := &LLMSummaryAgent{
		client: client,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func WithSummaryFallbackAgent(fallback SummaryAgent) LLMSummaryOption {
	return func(agent *LLMSummaryAgent) {
		agent.fallback = fallback
	}
}

func (a *LLMSummaryAgent) Summarize(input SummaryInput) (SummaryOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(input, errors.New("llm client is nil"))
	}

	response, err := a.client.CreateChatCompletion(context.Background(), llm.ChatRequest{
		Messages: toLLMMessages(BuildSummaryPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(input, fmt.Errorf("create chat completion: %w", err))
	}

	result, err := parseSummaryJSON(response.Content)
	if err != nil {
		return a.fallbackOrError(input, err)
	}
	result.Raw = response.Raw

	return result, nil
}

func (a *LLMSummaryAgent) fallbackOrError(input SummaryInput, err error) (SummaryOutput, error) {
	if a.fallback != nil {
		return a.fallback.Summarize(input)
	}

	return SummaryOutput{}, err
}

type summaryPayload struct {
	Summary           *string   `json:"summary"`
	MajorProblems     *[]string `json:"major_problems"`
	FrequentErrors    *[]string `json:"frequent_errors"`
	BetterExpressions *[]string `json:"better_expressions"`
	NextPracticePlan  *[]string `json:"next_practice_plan"`
}

func parseSummaryJSON(content string) (SummaryOutput, error) {
	var payload summaryPayload
	if err := decodeStrictJSONObject(content, &payload); err != nil {
		return SummaryOutput{}, fmt.Errorf("parse summary json: %w", err)
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

func requiredStringList(value *[]string, field string) ([]string, error) {
	if value == nil {
		return nil, errors.New(field + " is required")
	}

	result := make([]string, 0, len(*value))
	for i, item := range *value {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("%s[%d] is empty", field, i)
		}
		result = append(result, item)
	}

	return result, nil
}
