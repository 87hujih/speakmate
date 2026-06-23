package agent

import (
	"context"
	"fmt"
	"strings"

	"speakmate/internal/model"
)

// ConversationAgent 定义非流式 AI 对话回复能力。
type ConversationAgent interface {
	GenerateReply(ctx context.Context, input ConversationInput) (ConversationOutput, error)
}

// StreamingConversationAgent 定义可输出回复分片的 AI 对话能力。
type StreamingConversationAgent interface {
	ConversationAgent
	StreamReply(ctx context.Context, input ConversationInput, onDelta func(ConversationDelta) error) (ConversationOutput, error)
}

// ConversationInput 是生成 AI 回复所需的训练上下文。
type ConversationInput struct {
	Scenario    model.Scenario
	Session     model.Session
	History     []model.Message
	UserContent string
}

// ConversationDelta 表示 AI 流式回复中的一个增量片段。
type ConversationDelta struct {
	Content string
	Raw     string
}

// ConversationOutput 是 AI 对话回复的结构化输出。
type ConversationOutput struct {
	Reply    string
	Stage    string
	NextGoal string
	Raw      string
}

// HistoryMessages 将消息历史转换为 Agent 可消费的上下文列表。
func (input ConversationInput) HistoryMessages() []model.Message {
	if input.History != nil {
		return input.History
	}

	return input.Session.Messages
}

// StageNameForTurn 根据当前轮次推导训练阶段名称。
func StageNameForTurn(stages []model.ScenarioStage, turnIndex int) string {
	if len(stages) == 0 {
		return "general"
	}
	if turnIndex < 0 {
		turnIndex = 0
	}
	if turnIndex >= len(stages) {
		turnIndex = len(stages) - 1
	}

	return stages[turnIndex].Name
}

// stageDescriptionForTurn 根据轮次返回对应阶段说明。
func stageDescriptionForTurn(stages []model.ScenarioStage, turnIndex int) string {
	if len(stages) == 0 {
		return ""
	}
	if turnIndex < 0 {
		turnIndex = 0
	}
	if turnIndex >= len(stages) {
		turnIndex = len(stages) - 1
	}

	return strings.TrimSpace(stages[turnIndex].Description)
}

// NextGoalForTurn 根据场景阶段推导下一步训练目标。
func NextGoalForTurn(stages []model.ScenarioStage, turnIndex int) string {
	stage := StageNameForTurn(stages, turnIndex)
	description := stageDescriptionForTurn(stages, turnIndex)
	if description != "" {
		return fmt.Sprintf("Guide the user through %s: %s", stage, description)
	}
	if stage != "" && stage != "general" {
		return "Ask one specific follow-up for the " + stage + " stage"
	}

	return "Ask one specific follow-up that keeps the practice moving"
}
