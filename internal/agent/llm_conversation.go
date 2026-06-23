package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"speakmate/internal/infra/llm"
)

// LLMConversationAgent 使用 OpenAI-compatible LLM 生成训练对话回复。
type LLMConversationAgent struct {
	client          llm.Client
	streamingClient llm.StreamingClient
	fallback        ConversationAgent
}

// LLMConversationOption 用于配置 LLMConversationAgent。
type LLMConversationOption func(*LLMConversationAgent)

// NewLLMConversationAgent 创建并返回对应组件实例。
func NewLLMConversationAgent(client llm.Client, opts ...LLMConversationOption) *LLMConversationAgent {
	agent := &LLMConversationAgent{
		client: client,
	}
	if streamingClient, ok := client.(llm.StreamingClient); ok {
		agent.streamingClient = streamingClient
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// WithFallbackAgent 返回用于覆盖默认行为的配置选项。
func WithFallbackAgent(fallback ConversationAgent) LLMConversationOption {
	return func(agent *LLMConversationAgent) {
		agent.fallback = fallback
	}
}

// GenerateReply 调用 LLM 生成非流式 AI 回复，并在失败时按配置降级。
func (a *LLMConversationAgent) GenerateReply(ctx context.Context, input ConversationInput) (ConversationOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(ctx, input, errors.New("LLM 客户端不能为空"))
	}

	response, err := a.client.CreateChatCompletion(ctx, llm.ChatRequest{
		Messages: toLLMMessages(BuildConversationPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(ctx, input, fmt.Errorf("创建聊天补全失败：%w", err))
	}

	reply := cleanConversationReply(response.Content)
	if reply == "" {
		return a.fallbackOrError(ctx, input, errors.New("LLM 返回空回复"))
	}

	stageIndex := input.Session.TurnCount + 1
	return ConversationOutput{
		Reply:    reply,
		Stage:    StageNameForTurn(input.Scenario.Stages, stageIndex),
		NextGoal: NextGoalForTurn(input.Scenario.Stages, stageIndex),
		Raw:      response.Raw,
	}, nil
}

// StreamReply 调用 LLM 流式生成 AI 回复，并逐段回调给事件发布链路。
func (a *LLMConversationAgent) StreamReply(ctx context.Context, input ConversationInput, onDelta func(ConversationDelta) error) (ConversationOutput, error) {
	if a.streamingClient == nil {
		return a.fallbackStreamOrError(ctx, input, onDelta, errors.New("LLM 流式客户端不能为空"))
	}

	deltaCleaner := &conversationDeltaCleaner{}
	response, err := a.streamingClient.CreateChatCompletionStream(ctx, llm.ChatRequest{
		Messages: toLLMMessages(BuildConversationPrompt(input)),
	}, func(delta llm.ChatStreamDelta) error {
		if delta.Content == "" || onDelta == nil {
			return nil
		}
		content := deltaCleaner.Clean(delta.Content)
		if content == "" {
			return nil
		}

		return onDelta(ConversationDelta{
			Content: content,
			Raw:     delta.Raw,
		})
	})
	if err != nil {
		return a.fallbackStreamOrError(ctx, input, onDelta, fmt.Errorf("创建聊天补全流失败：%w", err))
	}

	if pending := deltaCleaner.Flush(); pending != "" && onDelta != nil {
		if err := onDelta(ConversationDelta{Content: pending}); err != nil {
			return ConversationOutput{}, err
		}
	}

	reply := cleanConversationReply(response.Content)
	if reply == "" {
		return a.fallbackStreamOrError(ctx, input, onDelta, errors.New("LLM 返回空回复"))
	}

	stageIndex := input.Session.TurnCount + 1
	return ConversationOutput{
		Reply:    reply,
		Stage:    StageNameForTurn(input.Scenario.Stages, stageIndex),
		NextGoal: NextGoalForTurn(input.Scenario.Stages, stageIndex),
		Raw:      response.Raw,
	}, nil
}

// fallbackOrError 在配置 fallback 时降级处理，否则返回原始错误。
func (a *LLMConversationAgent) fallbackOrError(ctx context.Context, input ConversationInput, err error) (ConversationOutput, error) {
	if a.fallback != nil {
		return a.fallback.GenerateReply(ctx, input)
	}

	return ConversationOutput{}, err
}

// fallbackStreamOrError 在流式回复失败时执行相同的降级策略。
func (a *LLMConversationAgent) fallbackStreamOrError(ctx context.Context, input ConversationInput, onDelta func(ConversationDelta) error, err error) (ConversationOutput, error) {
	if a.fallback == nil {
		return ConversationOutput{}, err
	}

	if streamingFallback, ok := a.fallback.(StreamingConversationAgent); ok {
		return streamingFallback.StreamReply(ctx, input, onDelta)
	}

	output, fallbackErr := a.fallback.GenerateReply(ctx, input)
	if fallbackErr != nil {
		return ConversationOutput{}, fallbackErr
	}
	if strings.TrimSpace(output.Reply) == "" || onDelta == nil {
		return output, nil
	}
	if err := onDelta(ConversationDelta{Content: output.Reply}); err != nil {
		return ConversationOutput{}, err
	}

	return output, nil
}

// toLLMMessages 将业务提示词转换为 LLM 请求消息。
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

type conversationDeltaCleaner struct {
	buffer      string
	passthrough bool
}

func (c *conversationDeltaCleaner) Clean(content string) string {
	if c.passthrough {
		return content
	}

	c.buffer += content
	cleaned, pending := stripLeadingReplyLabels(c.buffer)
	if pending {
		return ""
	}

	c.buffer = ""
	c.passthrough = true
	return cleaned
}

func (c *conversationDeltaCleaner) Flush() string {
	if c.passthrough || c.buffer == "" {
		return ""
	}

	pending := c.buffer
	c.buffer = ""
	c.passthrough = true
	return pending
}

func cleanConversationReply(content string) string {
	reply := strings.TrimSpace(content)
	for {
		cleaned, pending := stripLeadingReplyLabels(reply)
		if pending || cleaned == reply {
			return reply
		}
		reply = strings.TrimSpace(cleaned)
	}
}

func stripLeadingReplyLabels(content string) (string, bool) {
	reply := content
	stripped := false
	for {
		cleaned, changed, pending := stripOneLeadingReplyLabel(reply)
		if pending {
			if stripped {
				return "", false
			}
			return "", true
		}
		if !changed {
			return reply, false
		}
		stripped = true
		reply = cleaned
	}
}

func stripOneLeadingReplyLabel(content string) (string, bool, bool) {
	trimmed := strings.TrimLeftFunc(content, unicode.IsSpace)
	if trimmed == "" {
		return "", false, true
	}

	for _, pair := range []struct {
		open  string
		close string
	}{
		{open: "[", close: "]"},
		{open: "【", close: "】"},
	} {
		if !strings.HasPrefix(trimmed, pair.open) {
			continue
		}

		afterOpen := trimmed[len(pair.open):]
		closeIndex := strings.Index(afterOpen, pair.close)
		if closeIndex < 0 {
			if len(afterOpen) <= 96 {
				return "", false, true
			}
			return content, false, false
		}

		label := strings.TrimSpace(afterOpen[:closeIndex])
		if label == "" || len(label) > 96 {
			return content, false, false
		}

		rest := afterOpen[closeIndex+len(pair.close):]
		return strings.TrimLeftFunc(rest, unicode.IsSpace), true, false
	}

	return content, false, false
}
