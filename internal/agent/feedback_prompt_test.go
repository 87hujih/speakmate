package agent

import (
	"testing"

	"speakmate/internal/model"
)

func TestBuildCorrectionPromptIncludesStructuredJSONContract(t *testing.T) {
	input := CorrectionInput{
		Scenario: model.Scenario{
			Code:     "interview",
			Name:     "English interview",
			AIRole:   "technical interviewer",
			UserGoal: "explain project experience clearly",
			Stages: []model.ScenarioStage{
				{Name: "project experience", Description: "describe role and impact"},
			},
		},
		Session: model.Session{ID: 10, TurnCount: 0},
		UserMessage: model.Message{
			ID:        1001,
			SessionID: 10,
			Content:   "I am study computer science and I have did a project.",
			Stage:     "project experience",
		},
	}

	messages := BuildCorrectionPrompt(input)

	if len(messages) < 2 {
		t.Fatalf("prompt message count = %d, want at least 2", len(messages))
	}
	if messages[0].Role != "system" {
		t.Fatalf("first prompt role = %q, want system", messages[0].Role)
	}

	joined := joinPromptMessages(messages)
	assertContains(t, joined, "Return only valid JSON")
	assertContains(t, joined, `"message_id"`)
	assertContains(t, joined, `"original_text"`)
	assertContains(t, joined, `"corrected_text"`)
	assertContains(t, joined, `"errors"`)
	assertContains(t, joined, `"better_expressions"`)
	assertContains(t, joined, "grammar")
	assertContains(t, joined, "vocabulary")
	assertContains(t, joined, "expression")
	assertContains(t, joined, "structure")
	assertContains(t, joined, "scenario")
	assertContains(t, joined, "English interview")
	assertContains(t, joined, "technical interviewer")
	assertContains(t, joined, "I am study computer science")
}

func TestBuildScoringPromptIncludesStructuredJSONContractAndRubric(t *testing.T) {
	input := ScoringInput{
		Scenario: model.Scenario{
			Code:     "interview",
			Name:     "English interview",
			UserGoal: "explain project experience clearly",
			Rubric: []model.ScenarioRubric{
				{Name: "completion", Description: "answers the stage objective"},
			},
		},
		Session: model.Session{ID: 10, TurnCount: 1},
		UserMessage: model.Message{
			ID:        1001,
			SessionID: 10,
			Content:   "I am studying computer science.",
			Stage:     "project experience",
		},
		Correction: model.CorrectionResult{
			MessageID:     1001,
			SessionID:     10,
			OriginalText:  "I am study computer science.",
			CorrectedText: "I am studying computer science.",
			Errors: []model.CorrectionError{
				{Type: model.CorrectionErrorTypeGrammar, Span: "am study", Suggestion: "am studying", Explanation: "Use present participle after be."},
			},
		},
	}

	messages := BuildScoringPrompt(input)

	if len(messages) < 2 {
		t.Fatalf("prompt message count = %d, want at least 2", len(messages))
	}

	joined := joinPromptMessages(messages)
	assertContains(t, joined, "Return only valid JSON")
	assertContains(t, joined, `"fluency"`)
	assertContains(t, joined, `"grammar"`)
	assertContains(t, joined, `"expression"`)
	assertContains(t, joined, `"vocabulary"`)
	assertContains(t, joined, `"completion"`)
	assertContains(t, joined, `"total_score"`)
	assertContains(t, joined, "0 to 100")
	assertContains(t, joined, "answers the stage objective")
	assertContains(t, joined, "I am studying computer science.")
	assertContains(t, joined, "am study")
}
