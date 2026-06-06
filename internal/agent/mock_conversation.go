package agent

import "speakmate/internal/model"

// MockConversationAgent 基于场景和轮次返回稳定的本地 Mock 回复。
type MockConversationAgent struct{}

// NewMockConversationAgent 创建 Mock 对话 Agent。
func NewMockConversationAgent() *MockConversationAgent {
	return &MockConversationAgent{}
}

// Generate 根据场景编码和当前轮次选择一条英文追问。
func (a *MockConversationAgent) Generate(input ConversationInput) (ConversationOutput, error) {
	turnCount := input.TurnCount
	if turnCount < 0 {
		turnCount = 0
	}

	stage := stageNameForTurn(input.Scenario.Stages, turnCount+1)
	replies := repliesForScenario(input.Scenario.Code)
	nextGoals := nextGoalsForScenario(input.Scenario.Code)
	index := turnCount % len(replies)

	return ConversationOutput{
		Reply:    replies[index],
		Stage:    stage,
		NextGoal: nextGoals[index],
		Raw:      nil,
	}, nil
}

func repliesForScenario(code string) []string {
	switch code {
	case "interview":
		return []string{
			"That project sounds relevant. Could you explain your role in the project and one technical challenge you solved?",
			"Good. What trade-offs did you consider in the technical design, and why did you choose that approach?",
			"How did you work with teammates when the project became difficult?",
		}
	case "restaurant":
		return []string{
			"Sure. Are you looking for something light, spicy, vegetarian, or a house special today?",
			"Thanks. Do you have any allergies or ingredients you want us to avoid?",
			"Got it. Would you like to add a drink or appetizer before I confirm the order?",
		}
	case "meeting":
		return []string{
			"Thanks for the update. What is the main blocker we should discuss first?",
			"That makes sense. Which option do you recommend, and what is the trade-off?",
			"Before we wrap up, who should own the next action and by when?",
		}
	default:
		return []string{
			"Thanks for sharing that. Could you add one specific detail so we can continue the practice?",
		}
	}
}

func nextGoalsForScenario(code string) []string {
	switch code {
	case "interview":
		return []string{
			"ask user to describe personal project contribution",
			"ask user to explain technical design trade-offs",
			"ask user to describe teamwork under pressure",
		}
	case "restaurant":
		return []string{
			"ask user to express menu preferences",
			"ask user to clarify allergies or avoided ingredients",
			"ask user to confirm add-ons before ordering",
		}
	case "meeting":
		return []string{
			"ask user to identify the main blocker",
			"ask user to recommend an option with trade-offs",
			"ask user to assign the next action",
		}
	default:
		return []string{
			"ask user to add one specific detail",
		}
	}
}

func stageNameForTurn(stages []model.ScenarioStage, turnIndex int) string {
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
