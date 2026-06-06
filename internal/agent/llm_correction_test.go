package agent

import (
	"context"
	"testing"

	"speakmate/internal/infra/llm"
	"speakmate/internal/model"
)

func TestLLMCorrectionAgentUsesClientAndParsesStructuredJSON(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{
				"message_id": 1001,
				"original_text": "I am study computer science.",
				"corrected_text": "I am studying computer science.",
				"errors": [
					{
						"type": "grammar",
						"span": "am study",
						"suggestion": "am studying",
						"explanation": "be 动词后应接现在分词。"
					}
				],
				"better_expressions": ["I major in computer science."]
			}`,
			Raw: `{"raw":true}`,
		},
	}
	agent := NewLLMCorrectionAgent(client)

	output, err := agent.Correct(CorrectionInput{
		Scenario: model.Scenario{Code: "interview", Name: "English interview"},
		Session:  model.Session{ID: 10},
		UserMessage: model.Message{
			ID:        1001,
			SessionID: 10,
			Content:   "I am study computer science.",
		},
	})
	if err != nil {
		t.Fatalf("Correct returned error: %v", err)
	}

	result := output.Result
	if result.MessageID != 1001 {
		t.Fatalf("message_id = %d, want 1001", result.MessageID)
	}
	if result.SessionID != 10 {
		t.Fatalf("session_id = %d, want 10 from input context", result.SessionID)
	}
	if result.CorrectedText != "I am studying computer science." {
		t.Fatalf("corrected_text = %q, want parsed corrected text", result.CorrectedText)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("errors length = %d, want 1", len(result.Errors))
	}
	if result.Errors[0].Type != model.CorrectionErrorTypeGrammar {
		t.Fatalf("error type = %q, want grammar", result.Errors[0].Type)
	}
	if len(result.BetterExpressions) != 1 || result.BetterExpressions[0] != "I major in computer science." {
		t.Fatalf("better_expressions = %#v, want parsed suggestions", result.BetterExpressions)
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
	if !promptContains(client.requests[0].Messages, "I am study computer science.") {
		t.Fatalf("client prompt does not contain user message: %#v", client.requests[0].Messages)
	}
}

func TestLLMCorrectionAgentRejectsInvalidJSON(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{Content: "not json"},
	}
	agent := NewLLMCorrectionAgent(client)

	_, err := agent.Correct(validCorrectionInput())

	if err == nil {
		t.Fatal("Correct error = nil, want invalid JSON error")
	}
}

func TestLLMCorrectionAgentRejectsMissingRequiredFields(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{"message_id":1001,"original_text":"I am study computer science.","errors":[],"better_expressions":[]}`,
		},
	}
	agent := NewLLMCorrectionAgent(client)

	_, err := agent.Correct(validCorrectionInput())

	if err == nil {
		t.Fatal("Correct error = nil, want missing corrected_text error")
	}
}

func TestLLMCorrectionAgentRejectsInvalidErrorType(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{
				"message_id":1001,
				"original_text":"I am study computer science.",
				"corrected_text":"I am studying computer science.",
				"errors":[{"type":"tone","span":"am study","suggestion":"am studying","explanation":"bad type"}],
				"better_expressions":[]
			}`,
		},
	}
	agent := NewLLMCorrectionAgent(client)

	_, err := agent.Correct(validCorrectionInput())

	if err == nil {
		t.Fatal("Correct error = nil, want invalid error type error")
	}
}

func TestLLMCorrectionAgentFallsBackWhenClientFails(t *testing.T) {
	client := &fakeLLMClient{err: context.Canceled}
	agent := NewLLMCorrectionAgent(client, WithCorrectionFallbackAgent(NewMockCorrectionAgent()))

	output, err := agent.Correct(validCorrectionInput())
	if err != nil {
		t.Fatalf("Correct returned error: %v", err)
	}

	if output.Result.MessageID != 1001 {
		t.Fatalf("fallback message_id = %d, want 1001", output.Result.MessageID)
	}
	if output.Result.CorrectedText == "" {
		t.Fatal("fallback corrected_text is empty")
	}
}

func validCorrectionInput() CorrectionInput {
	return CorrectionInput{
		Scenario: model.Scenario{Code: "interview", Name: "English interview"},
		Session:  model.Session{ID: 10},
		UserMessage: model.Message{
			ID:        1001,
			SessionID: 10,
			Content:   "I am study computer science.",
		},
	}
}
