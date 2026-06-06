package service

import "speakmate/internal/model"

// ConversationGenerator 定义对话回复生成能力，后续可替换为真实 Agent。
type ConversationGenerator interface {
	GenerateReply(input ConversationInput) ConversationReply
}

// ConversationInput 是生成一条 AI 回复所需的上下文。
type ConversationInput struct {
	Scenario    model.Scenario
	Session     model.Session
	UserContent string
}

// ConversationReply 是 Mock Conversation 的生成结果。
type ConversationReply struct {
	Content string
	Stage   string
}

// MockConversationService 基于场景和轮次返回稳定的本地 Mock 回复。
type MockConversationService struct{}

// NewMockConversationService 创建 Mock 对话服务。
func NewMockConversationService() *MockConversationService {
	return &MockConversationService{}
}

// GenerateReply 根据场景编码和当前轮次选择一条英文追问。
func (s *MockConversationService) GenerateReply(input ConversationInput) ConversationReply {
	stage := stageNameForTurn(input.Scenario.Stages, input.Session.TurnCount+1)
	replies := repliesForScenario(input.Scenario.Code)
	index := input.Session.TurnCount % len(replies)

	return ConversationReply{
		Content: replies[index],
		Stage:   stage,
	}
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
