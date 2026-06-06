package agent

import "speakmate/internal/model"

// ConversationAgent 定义场景化对话回复生成能力。
type ConversationAgent interface {
	Generate(input ConversationInput) (ConversationOutput, error)
}

// ConversationInput 是生成一条 AI 回复所需的完整上下文。
type ConversationInput struct {
	Scenario    model.Scenario
	Session     model.Session
	History     []model.Message
	UserMessage string
	TurnCount   int
}

// ConversationOutput 是 Conversation Agent 的生成结果。
type ConversationOutput struct {
	Reply    string
	Stage    string
	NextGoal string
	Raw      any
}
