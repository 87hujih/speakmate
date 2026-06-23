package agent

import (
	"fmt"
	"strings"

	"speakmate/internal/model"
)

// PromptMessage 表示发送给 LLM 的单条提示词消息。
type PromptMessage struct {
	Role    string
	Content string
}

// BuildConversationPrompt 封装当前文件中的辅助处理逻辑。
func BuildConversationPrompt(input ConversationInput) []PromptMessage {
	currentStage := StageNameForTurn(input.Scenario.Stages, input.Session.TurnCount)
	currentStageDescription := stageDescriptionForTurn(input.Scenario.Stages, input.Session.TurnCount)
	role := strings.TrimSpace(input.Scenario.AIRole)
	if role == "" {
		role = "scenario conversation partner"
	}
	userGoal := strings.TrimSpace(input.Scenario.UserGoal)
	if userGoal == "" {
		userGoal = "complete the speaking practice naturally"
	}

	system := strings.Join([]string{
		"You are SpeakMate's scenario conversation agent.",
		"Stay in character and speak as the scenario AI role.",
		"Reply in English only, even if the user uses another language.",
		"Keep every reply to 1-3 sentences.",
		"Prioritize a natural follow-up over frequent grammar correction.",
		"Ask one natural follow-up that advances the user's training goal.",
		"Do not prefix replies with stage labels, brackets, speaker names, or metadata.",
		"Scenario: " + valueOrFallback(input.Scenario.Name, input.Scenario.Code, "general practice"),
		"AI role: " + role,
		"Training goal: " + userGoal,
		"Current stage: " + currentStage,
		"Stage objective: " + valueOrFallback(currentStageDescription, "continue the conversation"),
	}, "\n")

	messages := []PromptMessage{
		{Role: "system", Content: system},
	}

	for _, historyMessage := range input.HistoryMessages() {
		content := strings.TrimSpace(historyMessage.Content)
		if content == "" {
			continue
		}
		messages = append(messages, PromptMessage{
			Role:    promptRole(historyMessage.Role),
			Content: content,
		})
	}

	userContent := strings.TrimSpace(input.UserContent)
	messages = append(messages, PromptMessage{
		Role: "user",
		Content: fmt.Sprintf(
			"User's latest message: %s\nRespond as %s. Keep it English only, 1-3 sentences, and include a natural follow-up.",
			userContent,
			role,
		),
	})

	return messages
}

// promptRole 将消息角色转换为提示词中的说话人标签。
func promptRole(role model.MessageRole) string {
	switch role {
	case model.MessageRoleAI:
		return "assistant"
	case model.MessageRoleUser:
		return "user"
	default:
		return "user"
	}
}

// valueOrFallback 返回有效字符串或默认兜底值。
func valueOrFallback(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}
