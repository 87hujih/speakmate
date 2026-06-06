package service

import "speakmate/internal/model"

// ConversationInput 是 Mock Conversation 生成回复所需的上下文。
type ConversationInput struct {
	Scenario    model.Scenario
	TurnCount   int
	UserContent string
}

// ConversationReply 是 Mock Conversation 生成的一条 AI 回复。
type ConversationReply struct {
	Content string
	Stage   string
}

// ConversationService 根据场景和轮次生成场景化 Mock AI 回复。
type ConversationService struct {
	templates map[string][]string
}

// NewConversationService 创建 Mock Conversation 服务。
func NewConversationService() *ConversationService {
	return &ConversationService{
		templates: map[string][]string{
			"interview": {
				"That project sounds useful. Could you explain your role in the project and one technical challenge you solved?",
				"Let's go deeper into the technical design. What trade-off did you make, and why did you choose that approach?",
				"How did you work with teammates on this project, and what would you improve if you built it again?",
			},
			"restaurant": {
				"Sure. What kind of dishes do you prefer, and do you have any allergies or dietary restrictions?",
				"Would you like something spicy or mild, and do you want a drink with your meal?",
				"Let me confirm your order. Are there any extra requests before I send it to the kitchen?",
			},
			"meeting": {
				"Thanks for the update. Which priority do you think we should handle first, and what is the main reason?",
				"Could you clarify the impact of that blocker and what support you need from the team?",
				"Let's summarize the next steps. Who should own each action item after this meeting?",
			},
		},
	}
}

// Generate 返回一条与当前场景相关的 Mock AI 回复。
func (s *ConversationService) Generate(input ConversationInput) ConversationReply {
	templates := s.templates[input.Scenario.Code]
	if len(templates) == 0 {
		templates = []string{
			"Thanks for sharing that. Could you add one more detail so we can keep the practice focused?",
			"What would you like to clarify next in this situation?",
			"Could you summarize your main point in one clear sentence?",
		}
	}

	index := input.TurnCount % len(templates)
	return ConversationReply{
		Content: templates[index],
		Stage:   scenarioStageName(input.Scenario.Stages, input.TurnCount+1),
	}
}

func scenarioStageName(stages []model.ScenarioStage, index int) string {
	if len(stages) == 0 {
		return "对话"
	}
	if index < 0 {
		index = 0
	}
	if index >= len(stages) {
		index = len(stages) - 1
	}

	return stages[index].Name
}
