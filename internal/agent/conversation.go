package agent

import (
	"context"
	"fmt"
	"strings"

	"speakmate/internal/model"
)

type ConversationAgent interface {
	GenerateReply(ctx context.Context, input ConversationInput) (ConversationOutput, error)
}

type ConversationInput struct {
	Scenario    model.Scenario
	Session     model.Session
	History     []model.Message
	UserContent string
}

type ConversationOutput struct {
	Reply    string
	Stage    string
	NextGoal string
	Raw      string
}

func (input ConversationInput) HistoryMessages() []model.Message {
	if input.History != nil {
		return input.History
	}

	return input.Session.Messages
}

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
