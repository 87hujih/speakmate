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
