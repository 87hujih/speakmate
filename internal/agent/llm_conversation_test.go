package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"speakmate/internal/infra/llm"
	"speakmate/internal/model"
)

func TestLLMConversationAgentUsesClientAndReturnsStageAndNextGoal(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{Content: "That sounds interesting. What was your personal contribution?"},
	}
	agent := NewLLMConversationAgent(client)

	output, err := agent.GenerateReply(context.Background(), ConversationInput{
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
		Session:     model.Session{TurnCount: 0},
		UserContent: "I built a robot control project.",
	})
	if err != nil {
		t.Fatalf("GenerateReply returned error: %v", err)
	}

	if output.Reply != "That sounds interesting. What was your personal contribution?" {
		t.Fatalf("Reply = %q, want fake LLM content", output.Reply)
	}
	if output.Stage != "project experience" {
		t.Fatalf("Stage = %q, want project experience", output.Stage)
	}
	if output.NextGoal == "" {
		t.Fatal("NextGoal is empty")
	}
	if len(client.requests) != 1 {
		t.Fatalf("client request count = %d, want 1", len(client.requests))
	}
	if !promptContains(client.requests[0].Messages, "technical interviewer") {
		t.Fatalf("client prompt does not contain scenario role: %#v", client.requests[0].Messages)
	}
}

func TestLLMConversationAgentFallsBackToMockWhenClientFails(t *testing.T) {
	client := &fakeLLMClient{err: errors.New("upstream unavailable")}
	agent := NewLLMConversationAgent(client, WithFallbackAgent(NewMockConversationAgent()))

	output, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{Code: "restaurant", Stages: []model.ScenarioStage{{Name: "menu"}, {Name: "preference"}}},
		Session:  model.Session{TurnCount: 0},
	})
	if err != nil {
		t.Fatalf("GenerateReply returned error: %v", err)
	}

	if output.Reply == "" {
		t.Fatal("fallback reply is empty")
	}
	if output.Stage != "preference" {
		t.Fatalf("fallback Stage = %q, want preference", output.Stage)
	}
}

func TestLLMConversationAgentReturnsErrorWhenClientFailsWithoutFallback(t *testing.T) {
	client := &fakeLLMClient{err: errors.New("upstream unavailable")}
	agent := NewLLMConversationAgent(client)

	_, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{Code: "meeting"},
		Session:  model.Session{TurnCount: 0},
	})

	if err == nil {
		t.Fatal("GenerateReply error = nil, want error")
	}
}

func TestLLMConversationAgentRejectsEmptyLLMReply(t *testing.T) {
	client := &fakeLLMClient{response: llm.ChatResponse{Content: "   "}}
	agent := NewLLMConversationAgent(client)

	_, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{Code: "meeting"},
		Session:  model.Session{TurnCount: 0},
	})

	if err == nil {
		t.Fatal("GenerateReply error = nil, want error")
	}
}

type fakeLLMClient struct {
	response llm.ChatResponse
	err      error
	requests []llm.ChatRequest
}

func (c *fakeLLMClient) CreateChatCompletion(ctx context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
	c.requests = append(c.requests, request)
	if c.err != nil {
		return llm.ChatResponse{}, c.err
	}

	return c.response, nil
}

func promptContains(messages []llm.Message, value string) bool {
	for _, message := range messages {
		if strings.Contains(message.Content, value) {
			return true
		}
	}

	return false
}
