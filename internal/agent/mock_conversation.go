package agent

import "context"

type MockConversationAgent struct{}

func NewMockConversationAgent() *MockConversationAgent {
	return &MockConversationAgent{}
}

func (a *MockConversationAgent) GenerateReply(ctx context.Context, input ConversationInput) (ConversationOutput, error) {
	stageIndex := input.Session.TurnCount + 1
	stage := StageNameForTurn(input.Scenario.Stages, stageIndex)
	replies := repliesForScenario(input.Scenario.Code)
	index := input.Session.TurnCount % len(replies)

	return ConversationOutput{
		Reply:    replies[index],
		Stage:    stage,
		NextGoal: mockNextGoalForScenario(input.Scenario.Code, stage),
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

func mockNextGoalForScenario(code string, stage string) string {
	switch code {
	case "interview":
		return "Ask the candidate for concrete project details and personal contribution"
	case "restaurant":
		return "Clarify the guest's preference and move toward confirming the order"
	case "meeting":
		return "Clarify the main point and identify the next action"
	default:
		return "Ask one natural follow-up for the " + stage + " stage"
	}
}
