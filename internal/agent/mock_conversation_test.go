package agent_test

import (
	"strings"
	"testing"

	"speakmate/internal/agent"
	"speakmate/internal/model"
)

func TestMockConversationAgentGeneratesScenarioSpecificReplyAndStage(t *testing.T) {
	conversationAgent := agent.NewMockConversationAgent()

	output, err := conversationAgent.Generate(agent.ConversationInput{
		Scenario: model.Scenario{
			Code: "interview",
			Stages: []model.ScenarioStage{
				{Name: "自我介绍"},
				{Name: "项目经历"},
			},
		},
		Session:     model.Session{TurnCount: 0},
		UserMessage: "I built a robot control project.",
		TurnCount:   0,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if !strings.Contains(strings.ToLower(output.Reply), "project") {
		t.Fatalf("reply = %q, want interview project follow-up", output.Reply)
	}
	if output.Stage != "项目经历" {
		t.Fatalf("stage = %q, want 项目经历", output.Stage)
	}
	if output.NextGoal == "" {
		t.Fatal("next_goal is empty")
	}
	if output.Raw != nil {
		t.Fatalf("raw = %#v, want nil for mock output", output.Raw)
	}
}

func TestMockConversationAgentFallsBackForUnknownScenario(t *testing.T) {
	conversationAgent := agent.NewMockConversationAgent()

	output, err := conversationAgent.Generate(agent.ConversationInput{
		Scenario:    model.Scenario{Code: "unknown"},
		Session:     model.Session{TurnCount: 8},
		UserMessage: "Hello",
		TurnCount:   8,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if output.Reply == "" {
		t.Fatal("reply is empty")
	}
	if output.Stage != "general" {
		t.Fatalf("stage = %q, want general", output.Stage)
	}
	if output.NextGoal == "" {
		t.Fatal("next_goal is empty")
	}
}
