package model

// Scenario 表示一个可训练的口语练习场景。
type Scenario struct {
	// ID 是场景的内部数字标识。
	ID int `json:"id"`
	// Code 是场景的稳定业务编码，便于前端或后续 Agent 策略识别。
	Code string `json:"code"`
	// Name 是展示给用户的场景名称。
	Name string `json:"name"`
	// Description 是场景摘要，列表页使用。
	Description string `json:"description"`
	// Difficulty 表示场景难度，例如 easy、medium。
	Difficulty string `json:"difficulty"`
	// AIRole 描述 AI 在该场景中扮演的角色。
	AIRole string `json:"ai_role"`
	// UserGoal 描述用户完成该场景训练时应达成的目标。
	UserGoal string `json:"user_goal"`
	// OpeningMessage 是 AI 在训练开始时发送的开场白。
	OpeningMessage string `json:"opening_message"`
	// Stages 定义该场景的训练阶段。
	Stages []ScenarioStage `json:"stages"`
	// Rubric 定义该场景的评分维度。
	Rubric []ScenarioRubric `json:"rubric"`
}

// ScenarioStage 表示场景中的一个训练阶段。
type ScenarioStage struct {
	// Name 是阶段名称。
	Name string `json:"name"`
	// Description 描述该阶段要练习的表达任务。
	Description string `json:"description"`
}

// ScenarioRubric 表示场景中的一个评分维度。
type ScenarioRubric struct {
	// Name 是评分维度名称。
	Name string `json:"name"`
	// Description 描述该维度的评价重点。
	Description string `json:"description"`
}
