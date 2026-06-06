package agent_test

import (
	"testing"

	"speakmate/internal/agent"
	"speakmate/internal/model"
)

func TestMockScoringAgentReturnsStableScoreForCorrectionErrors(t *testing.T) {
	scoringAgent := agent.NewMockScoringAgent()

	output, err := scoringAgent.Score(agent.ScoringInput{
		Scenario: model.Scenario{Code: "interview"},
		Session:  model.Session{ID: 10},
		UserMessage: model.Message{
			ID:        1001,
			SessionID: 10,
			Content:   "I am study computer science and I have did a project.",
		},
		Correction: model.CorrectionResult{
			MessageID: 1001,
			SessionID: 10,
			Errors: []model.CorrectionError{
				{Type: model.CorrectionErrorTypeGrammar, Span: "am study"},
				{Type: model.CorrectionErrorTypeGrammar, Span: "have did"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	result := output.Result
	if result.MessageID != 1001 {
		t.Fatalf("message_id = %d, want 1001", result.MessageID)
	}
	if result.SessionID != 10 {
		t.Fatalf("session_id = %d, want 10", result.SessionID)
	}
	if result.Fluency != 75 {
		t.Fatalf("fluency = %d, want 75", result.Fluency)
	}
	if result.Grammar != 72 {
		t.Fatalf("grammar = %d, want 72", result.Grammar)
	}
	if result.Expression != 80 {
		t.Fatalf("expression = %d, want 80", result.Expression)
	}
	if result.Vocabulary != 76 {
		t.Fatalf("vocabulary = %d, want 76", result.Vocabulary)
	}
	if result.Completion != 85 {
		t.Fatalf("completion = %d, want 85", result.Completion)
	}
	if result.TotalScore != 77 {
		t.Fatalf("total_score = %d, want weighted total 77", result.TotalScore)
	}
	if result.Comment == "" {
		t.Fatal("comment is empty")
	}
	if output.Raw != nil {
		t.Fatalf("raw = %#v, want nil for mock output", output.Raw)
	}
}

func TestMockScoringAgentReturnsHigherStableScoreWithoutErrors(t *testing.T) {
	scoringAgent := agent.NewMockScoringAgent()

	output, err := scoringAgent.Score(agent.ScoringInput{
		Scenario: model.Scenario{Code: "meeting"},
		Session:  model.Session{ID: 20},
		UserMessage: model.Message{
			ID:        2001,
			SessionID: 20,
			Content:   "I recommend this option because it reduces risk.",
		},
		Correction: model.CorrectionResult{
			MessageID: 2001,
			SessionID: 20,
			Errors:    []model.CorrectionError{},
		},
	})
	if err != nil {
		t.Fatalf("Score returned error: %v", err)
	}

	result := output.Result
	if result.TotalScore != 85 {
		t.Fatalf("total_score = %d, want 85", result.TotalScore)
	}
	if result.Grammar != 88 {
		t.Fatalf("grammar = %d, want 88", result.Grammar)
	}
	if result.Comment == "" {
		t.Fatal("comment is empty")
	}
}
