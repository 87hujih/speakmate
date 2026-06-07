package agent

import (
	"strings"
	"testing"
	"time"

	"speakmate/internal/model"
)

func TestMockSummaryAgentBuildsStableReportSummary(t *testing.T) {
	agent := NewMockSummaryAgent()

	output, err := agent.Summarize(validSummaryInput())
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}

	if output.Summary == "" {
		t.Fatal("summary is empty")
	}
	if !strings.Contains(output.Summary, "77") {
		t.Fatalf("summary = %q, want total score context", output.Summary)
	}
	if len(output.MajorProblems) == 0 {
		t.Fatal("major_problems is empty")
	}
	if len(output.FrequentErrors) != 2 {
		t.Fatalf("frequent_errors length = %d, want 2", len(output.FrequentErrors))
	}
	if !strings.Contains(output.FrequentErrors[0], "am study -> am studying") {
		t.Fatalf("first frequent error = %q, want grammar suggestion", output.FrequentErrors[0])
	}
	if len(output.BetterExpressions) != 2 {
		t.Fatalf("better_expressions length = %d, want 2", len(output.BetterExpressions))
	}
	if !strings.Contains(output.BetterExpressions[0], "I major in computer science.") {
		t.Fatalf("first better expression = %q, want correction suggestion", output.BetterExpressions[0])
	}
	if len(output.NextPracticePlan) == 0 {
		t.Fatal("next_practice_plan is empty")
	}
}

func TestMockSummaryAgentAnchorsReportInConversationEvidence(t *testing.T) {
	agent := NewMockSummaryAgent()
	input := validSummaryInput()

	output, err := agent.Summarize(input)
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}

	if !strings.Contains(output.Summary, "I am study computer science") {
		t.Fatalf("summary = %q, want a concrete user utterance from the conversation", output.Summary)
	}
	if !containsStringWith(output.MajorProblems, "am study") {
		t.Fatalf("major_problems = %#v, want correction evidence from the user's wording", output.MajorProblems)
	}
	if !containsStringWith(output.NextPracticePlan, "have done") {
		t.Fatalf("next_practice_plan = %#v, want practice tied to a corrected expression", output.NextPracticePlan)
	}
}

func TestMockSummaryAgentReturnsFallbackSuggestionsWithoutCorrections(t *testing.T) {
	agent := NewMockSummaryAgent()
	input := validSummaryInput()
	input.Corrections = nil

	output, err := agent.Summarize(input)
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}

	if output.Summary == "" {
		t.Fatal("summary is empty")
	}
	if len(output.FrequentErrors) == 0 {
		t.Fatal("frequent_errors should contain fallback guidance")
	}
	if len(output.BetterExpressions) == 0 {
		t.Fatal("better_expressions should contain fallback guidance")
	}
	if len(output.NextPracticePlan) == 0 {
		t.Fatal("next_practice_plan should contain fallback guidance")
	}
}

func containsStringWith(values []string, fragment string) bool {
	for _, value := range values {
		if strings.Contains(value, fragment) {
			return true
		}
	}

	return false
}

func validSummaryInput() SummaryInput {
	createdAt := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)
	endedAt := createdAt.Add(3 * time.Minute)

	return SummaryInput{
		Scenario: model.Scenario{
			ID:         1,
			Code:       "interview",
			Name:       "英语面试",
			Difficulty: "medium",
			UserGoal:   "清晰介绍项目经历",
		},
		Session: model.Session{
			ID:        7,
			Status:    model.SessionStatusFinished,
			TurnCount: 2,
			CreatedAt: createdAt,
			EndedAt:   &endedAt,
		},
		Messages: []model.Message{
			{ID: 1, SessionID: 7, Role: model.MessageRoleUser, Content: "I am study computer science and I have did a project.", Stage: "self introduction"},
			{ID: 2, SessionID: 7, Role: model.MessageRoleAI, Content: "What was your role in the project?", Stage: "project experience"},
		},
		Corrections: []model.CorrectionResult{
			{
				MessageID:     1,
				SessionID:     7,
				OriginalText:  "I am study computer science and I have did a project.",
				CorrectedText: "I am studying computer science, and I have done a project.",
				Errors: []model.CorrectionError{
					{
						Type:        model.CorrectionErrorTypeGrammar,
						Span:        "am study",
						Suggestion:  "am studying",
						Explanation: "be 动词后应接现在分词。",
					},
					{
						Type:        model.CorrectionErrorTypeGrammar,
						Span:        "have did",
						Suggestion:  "have done",
						Explanation: "现在完成时中 have 后应接过去分词 done。",
					},
				},
				BetterExpressions: []string{
					"I major in computer science.",
					"I worked on a robotics project.",
				},
			},
		},
		Score: model.ScoreResult{
			MessageID:  1,
			SessionID:  7,
			Fluency:    75,
			Grammar:    72,
			Expression: 80,
			Vocabulary: 76,
			Completion: 85,
			TotalScore: 77,
			Comment:    "用户能够表达核心意思，但存在时态和动词形式错误。",
		},
	}
}
