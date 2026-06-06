package router

import (
	"testing"

	"speakmate/internal/agent"
	"speakmate/internal/config"
)

func TestNewConversationAgentReturnsMockWhenMockEnabled(t *testing.T) {
	got := NewConversationAgent(config.Config{
		LLM: config.LLMConfig{UseMock: true},
	})

	if _, ok := got.(*agent.MockConversationAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.MockConversationAgent", got)
	}
}

func TestNewConversationAgentReturnsMockWhenLLMConfigIncomplete(t *testing.T) {
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

	if _, ok := got.(*agent.MockConversationAgent); !ok {
		t.Fatalf("agent type = %T, want *agent.MockConversationAgent", got)
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
