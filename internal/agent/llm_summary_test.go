package agent

import (
	"context"
	"testing"

	"speakmate/internal/infra/llm"
)

func TestLLMSummaryAgentUsesClientAndParsesStructuredJSON(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{
				"summary": "本次面试训练能够说明项目背景，但需要加强动词形式。",
				"major_problems": ["动词形式不稳定"],
				"frequent_errors": ["am study -> am studying"],
				"better_expressions": ["I major in computer science."],
				"next_practice_plan": ["用 STAR 结构重写项目经历回答。"]
			}`,
			Raw: `{"raw":true}`,
		},
	}
	agent := NewLLMSummaryAgent(client)

	output, err := agent.Summarize(validSummaryInput())
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}

	if output.Summary != "本次面试训练能够说明项目背景，但需要加强动词形式。" {
		t.Fatalf("summary = %q, want parsed summary", output.Summary)
	}
	if len(output.MajorProblems) != 1 || output.MajorProblems[0] != "动词形式不稳定" {
		t.Fatalf("major_problems = %#v, want parsed major problem", output.MajorProblems)
	}
	if len(output.FrequentErrors) != 1 || output.FrequentErrors[0] != "am study -> am studying" {
		t.Fatalf("frequent_errors = %#v, want parsed frequent error", output.FrequentErrors)
	}
	if len(output.BetterExpressions) != 1 || output.BetterExpressions[0] != "I major in computer science." {
		t.Fatalf("better_expressions = %#v, want parsed expression", output.BetterExpressions)
	}
	if len(output.NextPracticePlan) != 1 || output.NextPracticePlan[0] != "用 STAR 结构重写项目经历回答。" {
		t.Fatalf("next_practice_plan = %#v, want parsed plan", output.NextPracticePlan)
	}
	if output.Raw != `{"raw":true}` {
		t.Fatalf("raw = %#v, want raw LLM response", output.Raw)
	}
	if len(client.requests) != 1 {
		t.Fatalf("client request count = %d, want 1", len(client.requests))
	}
	if !promptContains(client.requests[0].Messages, "Return only valid JSON") {
		t.Fatalf("client prompt does not contain JSON contract: %#v", client.requests[0].Messages)
	}
	if !promptContains(client.requests[0].Messages, "I am study computer science") {
		t.Fatalf("client prompt does not contain conversation history: %#v", client.requests[0].Messages)
	}
	if !promptContains(client.requests[0].Messages, `"total_score": 77`) {
		t.Fatalf("client prompt does not contain score JSON: %#v", client.requests[0].Messages)
	}
}

func TestLLMSummaryAgentRejectsInvalidJSON(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{Content: "not json"},
	}
	agent := NewLLMSummaryAgent(client)

	_, err := agent.Summarize(validSummaryInput())

	if err == nil {
		t.Fatal("Summarize error = nil, want invalid JSON error")
	}
}

func TestLLMSummaryAgentRejectsMissingRequiredFields(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{"summary":"ok","major_problems":[],"frequent_errors":[],"better_expressions":[]}`,
		},
	}
	agent := NewLLMSummaryAgent(client)

	_, err := agent.Summarize(validSummaryInput())

	if err == nil {
		t.Fatal("Summarize error = nil, want missing next_practice_plan error")
	}
}

func TestLLMSummaryAgentFallsBackWhenClientFails(t *testing.T) {
	client := &fakeLLMClient{err: context.Canceled}
	agent := NewLLMSummaryAgent(client, WithSummaryFallbackAgent(NewMockSummaryAgent()))

	output, err := agent.Summarize(validSummaryInput())
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}

	if output.Summary == "" {
		t.Fatal("fallback summary is empty")
	}
	if len(output.FrequentErrors) == 0 {
		t.Fatal("fallback frequent_errors is empty")
	}
}
