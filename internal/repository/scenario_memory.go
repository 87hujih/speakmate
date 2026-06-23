package repository

import (
	"errors"

	"speakmate/internal/model"
)

// ErrScenarioNotFound 表示内存仓库中没有找到对应场景。
var ErrScenarioNotFound = errors.New("未找到训练场景")

// MemoryScenarioRepository 使用内存数据提供场景查询能力。
type MemoryScenarioRepository struct {
	scenarios []model.Scenario
}

// NewMemoryScenarioRepository 创建带有内置训练场景的内存仓库。
func NewMemoryScenarioRepository() *MemoryScenarioRepository {
	return &MemoryScenarioRepository{
		scenarios: []model.Scenario{
			// 英语面试场景用于训练自我介绍、项目经历和技术追问。
			{
				ID:          1,
				Code:        "interview",
				Name:        "英语面试",
				Description: "练习自我介绍、项目经历和技术追问",
				Difficulty:  "medium",
				AIRole:      "技术面试官",
				UserGoal:    "清晰介绍自己的背景、项目经历和协作方式，并能回答常见技术追问。",
				OpeningMessage: "Hello, welcome to the interview. Could you start by briefly introducing yourself " +
					"and telling me about one project you are proud of?",
				Stages: []model.ScenarioStage{
					{
						Name:        "自我介绍",
						Description: "介绍教育背景、技术方向和当前目标。",
					},
					{
						Name:        "项目经历",
						Description: "说明项目目标、个人职责和关键成果。",
					},
					{
						Name:        "技术追问",
						Description: "回答实现细节、技术选择和问题排查过程。",
					},
					{
						Name:        "结尾提问",
						Description: "向面试官提出关于团队、岗位或项目的问题。",
					},
				},
				Rubric: []model.ScenarioRubric{
					{
						Name:        "流利度",
						Description: "回答是否连贯，是否能自然展开说明。",
					},
					{
						Name:        "语法准确度",
						Description: "时态、主谓一致和句子结构是否准确。",
					},
					{
						Name:        "表达自然度",
						Description: "是否使用面试场景中自然、得体的表达。",
					},
					{
						Name:        "场景完成度",
						Description: "是否完成自我介绍、项目说明和反问环节。",
					},
				},
			},
			// 餐厅点餐场景用于训练菜单询问、口味偏好和订单确认。
			{
				ID:          2,
				Code:        "restaurant",
				Name:        "餐厅点餐",
				Description: "练习询问菜单、表达偏好和处理特殊需求",
				Difficulty:  "easy",
				AIRole:      "餐厅服务员",
				UserGoal:    "完成点餐流程，清楚表达口味偏好、忌口和额外需求。",
				OpeningMessage: "Good evening. Welcome to our restaurant. Would you like to hear today's specials " +
					"before you order?",
				Stages: []model.ScenarioStage{
					{
						Name:        "询问菜单",
						Description: "询问推荐菜、套餐或今日特色。",
					},
					{
						Name:        "表达偏好",
						Description: "说明口味、饮食限制或过敏信息。",
					},
					{
						Name:        "确认订单",
						Description: "确认菜品、饮品和额外要求。",
					},
				},
				Rubric: []model.ScenarioRubric{
					{
						Name:        "清晰度",
						Description: "点餐需求是否表达明确。",
					},
					{
						Name:        "礼貌程度",
						Description: "是否使用礼貌、自然的服务场景表达。",
					},
					{
						Name:        "词汇匹配度",
						Description: "是否使用菜单、口味和数量相关词汇。",
					},
				},
			},
			// 工作会议场景用于训练进度同步、观点表达和分工确认。
			{
				ID:          3,
				Code:        "meeting",
				Name:        "工作会议",
				Description: "练习表达观点、澄清问题和总结结论",
				Difficulty:  "medium",
				AIRole:      "项目同事",
				UserGoal:    "在会议中清楚表达观点、回应问题，并推动形成下一步行动。",
				OpeningMessage: "Thanks for joining the meeting. Could you give us a quick update on your progress " +
					"and any blockers you are seeing?",
				Stages: []model.ScenarioStage{
					{
						Name:        "进度同步",
						Description: "说明当前进展、已完成事项和风险。",
					},
					{
						Name:        "观点表达",
						Description: "提出建议并解释理由。",
					},
					{
						Name:        "澄清确认",
						Description: "澄清问题、确认分工和下一步。",
					},
				},
				Rubric: []model.ScenarioRubric{
					{
						Name:        "结构清晰度",
						Description: "表达是否有清楚的先后顺序和重点。",
					},
					{
						Name:        "互动能力",
						Description: "是否能回应他人问题并主动澄清。",
					},
					{
						Name:        "场景词汇",
						Description: "是否使用会议、进度和协作相关表达。",
					},
					{
						Name:        "结论完整度",
						Description: "是否明确下一步行动和负责人。",
					},
				},
			},
		},
	}
}

// List 返回当前内置的全部训练场景。
func (r *MemoryScenarioRepository) List() []model.Scenario {
	scenarios := make([]model.Scenario, len(r.scenarios))
	// 返回切片副本，避免调用方直接修改仓库内部切片。
	copy(scenarios, r.scenarios)

	return scenarios
}

// FindByID 按数字 ID 查询单个训练场景。
func (r *MemoryScenarioRepository) FindByID(id int) (model.Scenario, error) {
	for _, scenario := range r.scenarios {
		if scenario.ID == id {
			return scenario, nil
		}
	}

	return model.Scenario{}, ErrScenarioNotFound
}
