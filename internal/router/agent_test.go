package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"speakmate/internal/agent"
	"speakmate/internal/config"
	"speakmate/internal/model"
)

func TestNewConversationAgentDoesNotReturnMockWhenMockEnabled(t *testing.T) {
	got := NewConversationAgent(config.Config{
		LLM: config.LLMConfig{UseMock: true},
	})

	if _, ok := got.(*agent.MockConversationAgent); !ok {
		return
	}

	t.Fatalf("agent type = %T, want non-mock conversation agent", got)
}

func TestNewConversationAgentFailsInsteadOfMockingWhenLLMConfigIncomplete(t *testing.T) {
	got := NewConversationAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        false,
			Provider:       "openai-compatible",
			BaseURL:        "",
			APIKey:         "",
			Model:          "",
			TimeoutSeconds: 30,
		},
	})

	if _, ok := got.(*agent.MockConversationAgent); ok {
		t.Fatalf("agent type = %T, want non-mock conversation agent", got)
	}
	output, err := got.GenerateReply(context.Background(), conversationInputForRouterTest())
	if err == nil {
		t.Fatal("GenerateReply error = nil, want configuration error")
	}
	if output.Reply != "" {
		t.Fatalf("reply = %q, want empty reply when LLM is unavailable", output.Reply)
	}
}

func TestNewConversationAgentReturnsLLMAgentWhenConfigured(t *testing.T) {
	got := NewConversationAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        false,
			Provider:       "openai-compatible",
			BaseURL:        "https://llm.example.com/v1",
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
	})

	if _, ok := got.(*agent.LLMConversationAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.LLMConversationAgent", got)
	}
}

func TestNewConversationAgentDoesNotFallbackToMockWhenLLMRequestFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "up实时事件流不可用", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	got := NewConversationAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        false,
			FallbackToMock: true,
			Provider:       "openai-compatible",
			BaseURL:        server.URL,
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
	})

	output, err := got.GenerateReply(context.Background(), conversationInputForRouterTest())
	if err == nil {
		t.Fatal("GenerateReply error = nil, want up事件流错误")
	}
	if output.Reply != "" {
		t.Fatalf("reply = %q, want empty reply instead of mock fallback", output.Reply)
	}
}

func conversationInputForRouterTest() agent.ConversationInput {
	return agent.ConversationInput{
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
	}
}

func TestNewCorrectionAgentReturnsLLMAgentWhenConfigured(t *testing.T) {
	got := NewCorrectionAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        false,
			Provider:       "openai-compatible",
			BaseURL:        "https://llm.example.com/v1",
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
		Feedback: config.FeedbackConfig{
			CorrectionUseMock: false,
		},
	})

	if _, ok := got.(*agent.LLMCorrectionAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.LLMCorrectionAgent", got)
	}
}

func TestNewCorrectionAgentReturnsMockWhenLLMMockEnabled(t *testing.T) {
	got := NewCorrectionAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        true,
			Provider:       "openai-compatible",
			BaseURL:        "https://llm.example.com/v1",
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
		Feedback: config.FeedbackConfig{
			CorrectionUseMock: false,
		},
	})

	if _, ok := got.(*agent.MockCorrectionAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.MockCorrectionAgent", got)
	}
}

func TestNewScoringAgentReturnsLLMAgentWhenConfigured(t *testing.T) {
	got := NewScoringAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        false,
			Provider:       "openai-compatible",
			BaseURL:        "https://llm.example.com/v1",
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
		Feedback: config.FeedbackConfig{
			ScoringUseMock: false,
		},
	})

	if _, ok := got.(*agent.LLMScoringAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.LLMScoringAgent", got)
	}
}

func TestNewScoringAgentReturnsMockWhenLLMMockEnabled(t *testing.T) {
	got := NewScoringAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        true,
			Provider:       "openai-compatible",
			BaseURL:        "https://llm.example.com/v1",
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
		Feedback: config.FeedbackConfig{
			ScoringUseMock: false,
		},
	})

	if _, ok := got.(*agent.MockScoringAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.MockScoringAgent", got)
	}
}

func TestNewSummaryAgentReturnsLLMAgentWhenConfigured(t *testing.T) {
	got := NewSummaryAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        false,
			Provider:       "openai-compatible",
			BaseURL:        "https://llm.example.com/v1",
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
		Feedback: config.FeedbackConfig{
			SummaryUseMock: false,
		},
	})

	if _, ok := got.(*agent.LLMSummaryAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.LLMSummaryAgent", got)
	}
}

func TestNewSummaryAgentReturnsMockWhenSummaryMockEnabled(t *testing.T) {
	got := NewSummaryAgent(config.Config{
		LLM: config.LLMConfig{
			UseMock:        false,
			Provider:       "openai-compatible",
			BaseURL:        "https://llm.example.com/v1",
			APIKey:         "test-key",
			Model:          "test-model",
			TimeoutSeconds: 30,
		},
		Feedback: config.FeedbackConfig{
			SummaryUseMock: true,
		},
	})

	if _, ok := got.(*agent.MockSummaryAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.MockSummaryAgent", got)
	}
}
