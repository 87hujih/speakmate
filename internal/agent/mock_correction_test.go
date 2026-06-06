package agent_test

import (
	"testing"

	"speakmate/internal/agent"
	"speakmate/internal/model"
)

func TestMockCorrectionAgentReturnsStableCorrections(t *testing.T) {
	correctionAgent := agent.NewMockCorrectionAgent()

	output, err := correctionAgent.Correct(agent.CorrectionInput{
		Scenario: model.Scenario{Code: "interview"},
		Session:  model.Session{ID: 10},
		UserMessage: model.Message{
			ID:        1001,
			SessionID: 10,
			Content:   "I am study computer science and I have did a project.",
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
		t.Fatalf("session_id = %d, want 10", result.SessionID)
	}
	if result.OriginalText != "I am study computer science and I have did a project." {
		t.Fatalf("original_text = %q, want input text", result.OriginalText)
	}
	if result.CorrectedText != "I am studying computer science, and I have done a project." {
		t.Fatalf("corrected_text = %q, want stable corrected text", result.CorrectedText)
	}
	if len(result.Errors) != 2 {
		t.Fatalf("errors length = %d, want 2", len(result.Errors))
	}
	if result.Errors[0].Type != model.CorrectionErrorTypeGrammar {
		t.Fatalf("first error type = %q, want grammar", result.Errors[0].Type)
	}
	if result.Errors[0].Span != "am study" {
		t.Fatalf("first error span = %q, want am study", result.Errors[0].Span)
	}
	if result.Errors[0].Suggestion != "am studying" {
		t.Fatalf("first error suggestion = %q, want am studying", result.Errors[0].Suggestion)
	}
	if result.Errors[0].Explanation == "" {
		t.Fatal("first error explanation is empty")
	}
	if result.Errors[1].Span != "have did" {
		t.Fatalf("second error span = %q, want have did", result.Errors[1].Span)
	}
	if len(result.BetterExpressions) != 2 {
		t.Fatalf("better_expressions length = %d, want 2", len(result.BetterExpressions))
	}
	if result.BetterExpressions[0] != "I major in computer science." {
		t.Fatalf("first better expression = %q, want interview expression", result.BetterExpressions[0])
	}
	if output.Raw != nil {
		t.Fatalf("raw = %#v, want nil for mock output", output.Raw)
	}
}

func TestMockCorrectionAgentReturnsNoErrorsForClearText(t *testing.T) {
	correctionAgent := agent.NewMockCorrectionAgent()

	output, err := correctionAgent.Correct(agent.CorrectionInput{
		Scenario: model.Scenario{Code: "meeting"},
		Session:  model.Session{ID: 20},
		UserMessage: model.Message{
			ID:        2001,
			SessionID: 20,
			Content:   "I recommend this option because it reduces risk.",
		},
	})
	if err != nil {
		t.Fatalf("Correct returned error: %v", err)
	}

	result := output.Result
	if result.CorrectedText != result.OriginalText {
		t.Fatalf("corrected_text = %q, want original text %q", result.CorrectedText, result.OriginalText)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("errors length = %d, want 0", len(result.Errors))
	}
	if len(result.BetterExpressions) != 2 {
		t.Fatalf("better_expressions length = %d, want 2", len(result.BetterExpressions))
	}
	if result.BetterExpressions[0] != "The main blocker is the API integration timeline." {
		t.Fatalf("first better expression = %q, want meeting expression", result.BetterExpressions[0])
	}
}
