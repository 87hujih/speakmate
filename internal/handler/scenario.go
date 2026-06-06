package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/response"
	"speakmate/internal/service"
)

const (
	// scenarioNotFoundCode 是场景不存在时的临时业务错误码。
	scenarioNotFoundCode = 1001
	// invalidScenarioID 是 URL 中场景 ID 非法时的临时业务错误码。
	invalidScenarioID = 1002
)

// ScenarioService 定义 Handler 依赖的场景业务能力。
type ScenarioService interface {
	ListScenarios() []model.Scenario
	GetScenario(id int) (model.Scenario, error)
}

// ScenarioHandler 负责处理 Scenario API 的 HTTP 请求。
type ScenarioHandler struct {
	service ScenarioService
}

// NewScenarioHandler 创建 Scenario API Handler。
func NewScenarioHandler(service ScenarioService) *ScenarioHandler {
	return &ScenarioHandler{
		service: service,
	}
}

// List 返回场景列表，只暴露列表页需要的摘要字段。
func (h *ScenarioHandler) List(c *gin.Context) {
	response.Success(c, toScenarioSummaries(h.service.ListScenarios()))
}

// Detail 返回单个场景详情，包含角色、目标、开场白、阶段和评分维度。
func (h *ScenarioHandler) Detail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, invalidScenarioID, "invalid scenario id")
		return
	}

	// Handler 只处理 HTTP 参数与响应转换，场景查询交给 Service。
	scenario, err := h.service.GetScenario(id)
	if err != nil {
		if errors.Is(err, service.ErrScenarioNotFound) {
			response.Error(c, http.StatusNotFound, scenarioNotFoundCode, "scenario not found")
			return
		}

		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
		return
	}

	response.Success(c, scenario)
}

// scenarioSummary 是场景列表接口返回的摘要结构。
type scenarioSummary struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
}

// toScenarioSummaries 把完整场景模型转换成列表响应模型。
func toScenarioSummaries(scenarios []model.Scenario) []scenarioSummary {
	summaries := make([]scenarioSummary, 0, len(scenarios))
	for _, scenario := range scenarios {
		summaries = append(summaries, scenarioSummary{
			ID:          scenario.ID,
			Code:        scenario.Code,
			Name:        scenario.Name,
			Description: scenario.Description,
			Difficulty:  scenario.Difficulty,
		})
	}

	return summaries
}
