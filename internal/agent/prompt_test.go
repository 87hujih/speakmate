package agent

import (
	"strings"
	"testing"
	"time"

	"speakmate/internal/model"
)

func TestBuildConversationPromptIncludesScenarioStageHistoryAndUserInput(t *testing.T) {
	input := ConversationInput{
		Scenario: model.Scenario{
			Code:     "interview",
			Name:     "英语面试",
			AIRole:   "technical interviewer",
			UserGoal: "explain project experience clearly",
			Stages: []model.ScenarioStage{
				{Name: "self introduction", Description: "introduce background"},
				{Name: "project experience", Description: "describe role and impact"},
			},
		},
		Session: model.Session{
			TurnCount: 1,
			Messages: []model.Message{
				{Role: model.MessageRoleUser, Content: "I study computer science.", Stage: "self introduction", CreatedAt: time.Now()},
				{Role: model.MessageRoleAI, Content: "Could you tell me about a project?", Stage: "project experience", CreatedAt: time.Now()},
			},
		},
		UserContent: "I built a robot control project.",
	}

	messages := BuildConversationPrompt(input)

	if len(messages) < 4 {
		t.Fatalf("prompt message count = %d, want at least 4", len(messages))
	}
	if messages[0].Role != "system" {
		t.Fatalf("first prompt role = %q, want system", messages[0].Role)
	}

	joined := joinPromptMessages(messages)
	assertContains(t, joined, "英语面试")
	assertContains(t, joined, "technical interviewer")
	assertContains(t, joined, "explain project experience clearly")
	assertContains(t, joined, "project experience")
	assertContains(t, joined, "describe role and impact")
	assertContains(t, joined, "I study computer science.")
	assertContains(t, joined, "Could you tell me about a project?")
	assertContains(t, joined, "I built a robot control project.")
	assertContains(t, joined, "English only")
	assertContains(t, joined, "1-3 sentences")
	assertContains(t, joined, "natural follow-up")
}

func TestBuildConversationPromptKeepsStageMetadataOutOfHistoryText(t *testing.T) {
	input := ConversationInput{
		Scenario: model.Scenario{
			Code:   "meeting",
			AIRole: "project manager",
			Stages: []model.ScenarioStage{
				{Name: "进度同步"},
				{Name: "澄清确认"},
			},
		},
		Session: model.Session{
			TurnCount: 1,
			Messages: []model.Message{
				{
					Role:      model.MessageRoleAI,
					Content:   "Could you confirm the owner and next step?",
					Stage:     "澄清确认",
					CreatedAt: time.Now(),
				},
			},
		},
		UserContent: "Alice will own the follow-up.",
	}

	messages := BuildConversationPrompt(input)

	for _, message := range messages {
		if message.Role == "assistant" && strings.Contains(message.Content, "Could you confirm") {
			if message.Content != "Could you confirm the owner and next step?" {
				t.Fatalf("history content = %q, want plain conversation text", message.Content)
			}
			return
		}
	}

	t.Fatalf("assistant history message not found: %#v", messages)
}

func TestStageNameForTurnFallsBackToLastStage(t *testing.T) {
	stages := []model.ScenarioStage{
		{Name: "first"},
		{Name: "second"},
	}

	if got := StageNameForTurn(stages, 10); got != "second" {
		t.Fatalf("StageNameForTurn() = %q, want second", got)
	}
}

func joinPromptMessages(messages []PromptMessage) string {
	var b strings.Builder
	for _, message := range messages {
		b.WriteString(message.Role)
		b.WriteString(": ")
		b.WriteString(message.Content)
		b.WriteString("\n")
	}

	return b.String()
}

func assertContains(t *testing.T, value string, want string) {
	t.Helper()

	if !strings.Contains(value, want) {
		t.Fatalf("value does not contain %q:\n%s", want, value)
	}
}
