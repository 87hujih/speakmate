package agent

import (
	"context"
	"errors"
	"strings"
)

// UnavailableConversationAgent 在真实 LLM 不可用时返回明确错误。
type UnavailableConversationAgent struct {
	err error
}

// NewUnavailableConversationAgent 创建并返回对应组件实例。
func NewUnavailableConversationAgent(reason string) *UnavailableConversationAgent {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "对话 LLM 不可用"
	}

	return &UnavailableConversationAgent{err: errors.New(reason)}
}

// GenerateReply 封装当前文件中的辅助处理逻辑。
func (a *UnavailableConversationAgent) GenerateReply(ctx context.Context, input ConversationInput) (ConversationOutput, error) {
	return ConversationOutput{}, a.err
}
