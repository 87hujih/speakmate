package agent_test

import (
	"context"
	"strings"
	"testing"

	"speakmate/internal/agent"
	"speakmate/internal/model"
)

func TestMockConversationAgentGeneratesScenarioSpecificReplyAndStage(t *testing.T) {
	conversationAgent := agent.NewMockConversationAgent()

	output, err := conversationAgent.GenerateReply(context.Background(), agent.ConversationInput{
		Scenario: model.Scenario{
			Code: "interview",
			Stages: []model.ScenarioStage{
				{Name: "自我介绍"},
				{Name: "项目经历"},
			},
		},
		Session:     model.Session{TurnCount: 0},
		UserContent: "I built a robot control project.",
	})
	if err != nil {
		t.Fatalf("GenerateReply returned error: %v", err)
	}

	if !strings.Contains(strings.ToLower(output.Reply), "project") {
		t.Fatalf("reply = %q, want interview project follow-up", output.Reply)
	}
	if output.Stage != "项目经历" {
		t.Fatalf("stage = %q, want 项目经历", output.Stage)
	}
	if output.NextGoal != "ask user to describe personal project contribution" {
		t.Fatalf("next_goal = %q, want project contribution goal", output.NextGoal)
	}
	if output.Raw != "" {
		t.Fatalf("raw = %q, want empty for mock output", output.Raw)
	}
}

func TestMockConversationAgentFallsBackForUnknownScenario(t *testing.T) {
	conversationAgent := agent.NewMockConversationAgent()

	output, err := conversationAgent.GenerateReply(context.Background(), agent.ConversationInput{
		Scenario:    model.Scenario{Code: "unknown"},
		Session:     model.Session{TurnCount: 8},
		UserContent: "Hello",
	})
	if err != nil {
		t.Fatalf("GenerateReply returned error: %v", err)
	}

	if output.Reply == "" {
		t.Fatal("reply is empty")
	}
	if output.Stage != "general" {
		t.Fatalf("stage = %q, want general", output.Stage)
	}
	if output.NextGoal != "ask user to add one specific detail" {
		t.Fatalf("next_goal = %q, want fallback detail goal", output.NextGoal)
	}
}

func TestMockConversationAgentStreamsFakeReplyChunks(t *testing.T) {
	conversationAgent := agent.NewMockConversationAgent()

	var chunks []string
	output, err := conversationAgent.StreamReply(context.Background(), agent.ConversationInput{
		Scenario: model.Scenario{
			Code: "interview",
			Stages: []model.ScenarioStage{
				{Name: "自我介绍"},
				{Name: "项目经历"},
			},
		},
		Session: model.Session{TurnCount: 0},
	}, func(delta agent.ConversationDelta) error {
		chunks = append(chunks, delta.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamReply returned error: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("chunks length = %d, want multiple fake streaming chunks", len(chunks))
	}
	if strings.Join(chunks, "") != output.Reply {
		t.Fatalf("joined chunks = %q, want full reply %q", strings.Join(chunks, ""), output.Reply)
	}
}
