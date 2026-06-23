package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"speakmate/internal/response"
	"speakmate/internal/service"
)

const invalidHistoryRequestCode = 6001

// HistoryService 定义 History Handler 依赖的业务能力。
type HistoryService interface {
	ListSessions(input service.HistoryListInput) (service.HistoryListResult, error)
}

// HistoryHandler 负责处理训练历史 API。
type HistoryHandler struct {
	service HistoryService
}

// NewHistoryHandler 创建 History API Handler。
func NewHistoryHandler(service HistoryService) *HistoryHandler {
	return &HistoryHandler{service: service}
}

// List 查询训练历史列表。
func (h *HistoryHandler) List(c *gin.Context) {
	page, pageSize, ok := parseHistoryPagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListSessions(service.HistoryListInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeHistoryError(c, err)
		return
	}

	response.Success(c, toHistoryListResponse(result))
}

// ListByUser 查询指定用户的训练历史列表。
func (h *HistoryHandler) ListByUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || userID <= 0 {
		response.Error(c, http.StatusBadRequest, invalidHistoryRequestCode, "历史记录请求无效")
		return
	}
	page, pageSize, ok := parseHistoryPagination(c)
	if !ok {
		return
	}

	result, err := h.service.ListSessions(service.HistoryListInput{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		writeHistoryError(c, err)
		return
	}

	response.Success(c, toHistoryListResponse(result))
}

// historyListResponse 是历史记录列表接口返回结构。
type historyListResponse struct {
	Items    []historyItemResponse `json:"items"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int                   `json:"total"`
}

// historyItemResponse 是历史记录列表中的单条响应结构。
type historyItemResponse struct {
	SessionID    int             `json:"session_id"`
	SessionNo    string          `json:"session_no"`
	UserID       int             `json:"user_id"`
	Scenario     scenarioSummary `json:"scenario"`
	Status       string          `json:"status"`
	TurnCount    int             `json:"turn_count"`
	TotalScore   *int            `json:"total_score"`
	ReportStatus string          `json:"report_status"`
	CreatedAt    string          `json:"created_at"`
	EndedAt      *string         `json:"ended_at"`
}

// parseHistoryPagination 解析历史列表分页参数。
func parseHistoryPagination(c *gin.Context) (int, int, bool) {
	page, ok := parsePositiveQueryInt(c, "page")
	if !ok {
		return 0, 0, false
	}
	pageSize, ok := parsePositiveQueryInt(c, "page_size")
	if !ok {
		return 0, 0, false
	}

	return page, pageSize, true
}

// parsePositiveQueryInt 解析查询参数中的正整数。
func parsePositiveQueryInt(c *gin.Context, key string) (int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return 0, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		response.Error(c, http.StatusBadRequest, invalidHistoryRequestCode, "历史记录请求无效")
		return 0, false
	}

	return value, true
}

// toHistoryListResponse 将历史业务结果转换为 HTTP 响应结构。
func toHistoryListResponse(result service.HistoryListResult) historyListResponse {
	items := make([]historyItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		reportStatus := "not_generated"
		if item.ReportGenerated {
			reportStatus = "generated"
		}
		items = append(items, historyItemResponse{
			SessionID: item.Session.ID,
			SessionNo: item.Session.SessionNo,
			UserID:    item.Session.UserID,
			Scenario: scenarioSummary{
				ID:          item.Scenario.ID,
				Code:        item.Scenario.Code,
				Name:        item.Scenario.Name,
				Description: item.Scenario.Description,
				Difficulty:  item.Scenario.Difficulty,
			},
			Status:       string(item.Session.Status),
			TurnCount:    item.Session.TurnCount,
			TotalScore:   item.TotalScore,
			ReportStatus: reportStatus,
			CreatedAt:    formatTime(item.Session.CreatedAt),
			EndedAt:      formatOptionalTime(item.Session.EndedAt),
		})
	}

	return historyListResponse{
		Items:    items,
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}
}

// writeHistoryError 将历史查询错误转换为统一 HTTP 响应。
func writeHistoryError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidHistoryRequest) {
		response.Error(c, http.StatusBadRequest, invalidHistoryRequestCode, "历史记录请求无效")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "服务器内部错误")
}
