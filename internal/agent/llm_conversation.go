package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"speakmate/internal/infra/llm"
)

type LLMConversationAgent struct {
	client   llm.Client
	fallback ConversationAgent
}

type LLMConversationOption func(*LLMConversationAgent)

func NewLLMConversationAgent(client llm.Client, opts ...LLMConversationOption) *LLMConversationAgent {
	agent := &LLMConversationAgent{
		client: client,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func WithFallbackAgent(fallback ConversationAgent) LLMConversationOption {
	return func(agent *LLMConversationAgent) {
		agent.fallback = fallback
	}
}

func (a *LLMConversationAgent) GenerateReply(ctx context.Context, input ConversationInput) (ConversationOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(ctx, input, errors.New("llm client is nil"))
	}

	response, err := a.client.CreateChatCompletion(ctx, llm.ChatRequest{
		Messages: toLLMMessages(BuildConversationPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(ctx, input, fmt.Errorf("create chat completion: %w", err))
	}

	reply := strings.TrimSpace(response.Content)
	if reply == "" {
		return a.fallbackOrError(ctx, input, errors.New("llm returned empty reply"))
	}

	stageIndex := input.Session.TurnCount + 1
	return ConversationOutput{
		Reply:    reply,
		Stage:    StageNameForTurn(input.Scenario.Stages, stageIndex),
		NextGoal: NextGoalForTurn(input.Scenario.Stages, stageIndex),
		Raw:      response.Raw,
	}, nil
}

func (a *LLMConversationAgent) fallbackOrError(ctx context.Context, input ConversationInput, err error) (ConversationOutput, error) {
	if a.fallback != nil {
		return a.fallback.GenerateReply(ctx, input)
	}

	return ConversationOutput{}, err
}

func toLLMMessages(messages []PromptMessage) []llm.Message {
	converted := make([]llm.Message, 0, len(messages))
	for _, message := range messages {
		converted = append(converted, llm.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	return converted
}
