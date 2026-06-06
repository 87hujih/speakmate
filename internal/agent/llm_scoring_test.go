package agent

import (
	"context"
	"testing"

	"speakmate/internal/infra/llm"
	"speakmate/internal/model"
)

func TestLLMScoringAgentUsesClientAndParsesStructuredJSON(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{
				"message_id": 1001,
				"fluency": 75,
				"grammar": 72,
				"expression": 80,
				"vocabulary": 76,
				"completion": 85,
				"total_score": 77,
				"comment": "用户能够表达核心意思，但存在时态和动词形式错误。"
			}`,
			Raw: `{"raw":true}`,
		},
	}
	agent := NewLLMScoringAgent(client)

	output, err := agent.Score(validScoringInput())
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	result := output.Result
	if result.MessageID != 1001 {
		t.Fatalf("message_id = %d, want 1001", result.MessageID)
	}
	if result.SessionID != 10 {
		t.Fatalf("session_id = %d, want 10 from input context", result.SessionID)
	}
	if result.Fluency != 75 || result.Grammar != 72 || result.Expression != 80 || result.Vocabulary != 76 || result.Completion != 85 {
		t.Fatalf("score dimensions = %+v, want parsed score dimensions", result)
	}
	if result.TotalScore != 77 {
		t.Fatalf("total_score = %d, want 77", result.TotalScore)
	}
	if result.Comment == "" {
		t.Fatal("comment is empty")
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
	if !promptContains(client.requests[0].Messages, "am study") {
		t.Fatalf("client prompt does not contain correction context: %#v", client.requests[0].Messages)
	}
}

func TestLLMScoringAgentRejectsInvalidJSON(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{Content: "not json"},
	}
	agent := NewLLMScoringAgent(client)

	_, err := agent.Score(validScoringInput())

	if err == nil {
		t.Fatal("Score error = nil, want invalid JSON error")
	}
}

func TestLLMScoringAgentRejectsMissingRequiredFields(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{"message_id":1001,"fluency":75,"grammar":72,"expression":80,"vocabulary":76,"completion":85,"total_score":77}`,
		},
	}
	agent := NewLLMScoringAgent(client)

	_, err := agent.Score(validScoringInput())

	if err == nil {
		t.Fatal("Score error = nil, want missing comment error")
	}
}

func TestLLMScoringAgentRejectsScoreOutOfRange(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{
			Content: `{
				"message_id":1001,
				"fluency":101,
				"grammar":72,
				"expression":80,
				"vocabulary":76,
				"completion":85,
				"total_score":77,
				"comment":"out of range"
			}`,
		},
	}
	agent := NewLLMScoringAgent(client)

	_, err := agent.Score(validScoringInput())

	if err == nil {
		t.Fatal("Score error = nil, want out-of-range score error")
	}
}

func TestLLMScoringAgentFallsBackWhenClientFails(t *testing.T) {
	client := &fakeLLMClient{err: context.Canceled}
	agent := NewLLMScoringAgent(client, WithScoringFallbackAgent(NewMockScoringAgent()))

	output, err := agent.Score(validScoringInput())
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	if output.Result.MessageID != 1001 {
		t.Fatalf("fallback message_id = %d, want 1001", output.Result.MessageID)
	}
	if output.Result.TotalScore == 0 {
		t.Fatal("fallback total_score is zero")
	}
}

func validScoringInput() ScoringInput {
	return ScoringInput{
		Scenario: model.Scenario{Code: "interview", Name: "English interview"},
		Session:  model.Session{ID: 10},
		UserMessage: model.Message{
			ID:        1001,
			SessionID: 10,
			Content:   "I am study computer science.",
		},
		Correction: model.CorrectionResult{
			MessageID:     1001,
			SessionID:     10,
			OriginalText:  "I am study computer science.",
			CorrectedText: "I am studying computer science.",
			Errors: []model.CorrectionError{
				{
					Type:        model.CorrectionErrorTypeGrammar,
					Span:        "am study",
					Suggestion:  "am studying",
					Explanation: "be 动词后应接现在分词。",
				},
			},
		},
	}
}
