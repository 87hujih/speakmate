package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/response"
	"speakmate/internal/service"
)

// 当前模块使用的业务错误码和事件常量。
const (
	invalidSessionRequestCode  = 2001
	invalidSessionIDCode       = 2002
	sessionNotFoundCode        = 2003
	sessionAlreadyFinishedCode = 2004
)

// SessionService 定义 Handler 依赖的 Session 业务能力。
type SessionService interface {
	CreateSession(input service.CreateSessionInput) (service.CreateSessionResult, error)
	GetSession(id int) (service.GetSessionResult, error)
	FinishSession(id int) (model.Session, error)
}

// SessionHandler 负责处理 Session API 的 HTTP 请求。
type SessionHandler struct {
	service SessionService
}

// NewSessionHandler 创建 Session API Handler。
func NewSessionHandler(service SessionService) *SessionHandler {
	return &SessionHandler{
		service: service,
	}
}

// Create 创建新的训练 Session。
func (h *SessionHandler) Create(c *gin.Context) {
	var req createSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ScenarioID <= 0 {
		response.Error(c, http.StatusBadRequest, invalidSessionRequestCode, "训练请求无效")
		return
	}
	if req.UserID <= 0 {
		req.UserID = 1
	}

	result, err := h.service.CreateSession(service.CreateSessionInput{
		ScenarioID: req.ScenarioID,
		UserID:     req.UserID,
	})
	if err != nil {
		writeSessionError(c, err)
		return
	}

	response.Success(c, createSessionResponse{
		SessionID:      result.Session.ID,
		SessionNo:      result.Session.SessionNo,
		ScenarioID:     result.Session.ScenarioID,
		Status:         string(result.Session.Status),
		OpeningMessage: result.OpeningMessage,
	})
}

// Detail 查询训练 Session 当前状态和基础信息。
func (h *SessionHandler) Detail(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}

	result, err := h.service.GetSession(id)
	if err != nil {
		writeSessionError(c, err)
		return
	}

	response.Success(c, toSessionDetailResponse(result))
}

// Finish 结束一次训练 Session。
func (h *SessionHandler) Finish(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}

	session, err := h.service.FinishSession(id)
	if err != nil {
		writeSessionError(c, err)
		return
	}

	response.Success(c, finishSessionResponse{
		SessionID: session.ID,
		Status:    string(session.Status),
		TurnCount: session.TurnCount,
		EndedAt:   formatTime(*session.EndedAt),
	})
}

// createSessionRequest 是创建训练 Session 的请求结构。
type createSessionRequest struct {
	ScenarioID int `json:"scenario_id"`
	UserID     int `json:"user_id"`
}

// createSessionResponse 是创建训练 Session 的返回结构。
type createSessionResponse struct {
	SessionID      int    `json:"session_id"`
	SessionNo      string `json:"session_no"`
	ScenarioID     int    `json:"scenario_id"`
	Status         string `json:"status"`
	OpeningMessage string `json:"opening_message"`
}

// sessionDetailResponse 是训练 Session 详情返回结构。
type sessionDetailResponse struct {
	SessionID int               `json:"session_id"`
	SessionNo string            `json:"session_no"`
	Scenario  scenarioSummary   `json:"scenario"`
	Status    string            `json:"status"`
	TurnCount int               `json:"turn_count"`
	Messages  []messageResponse `json:"messages"`
	CreatedAt string            `json:"created_at"`
	EndedAt   *string           `json:"ended_at"`
}

// messageResponse 是 HTTP API 中的消息返回结构。
type messageResponse struct {
	ID        int    `json:"id"`
	SessionID int    `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Stage     string `json:"stage"`
	CreatedAt string `json:"created_at"`
}

// finishSessionResponse 是结束训练 Session 的返回结构。
type finishSessionResponse struct {
	SessionID int    `json:"session_id"`
	Status    string `json:"status"`
	TurnCount int    `json:"turn_count"`
	EndedAt   string `json:"ended_at"`
}

// parsePositiveSessionID 从路径参数解析训练 Session ID。
func parsePositiveSessionID(c *gin.Context) (int, bool) {
	rawID := c.Param("id")
	if rawID == "" {
		rawID = c.Param("session_id")
	}

	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, invalidSessionIDCode, "训练 ID 无效")
		return 0, false
	}

	return id, true
}

// toSessionDetailResponse 将 Session 详情转换为 HTTP 响应结构。
func toSessionDetailResponse(result service.GetSessionResult) sessionDetailResponse {
	return sessionDetailResponse{
		SessionID: result.Session.ID,
		SessionNo: result.Session.SessionNo,
		Scenario: scenarioSummary{
			ID:          result.Scenario.ID,
			Code:        result.Scenario.Code,
			Name:        result.Scenario.Name,
			Description: result.Scenario.Description,
			Difficulty:  result.Scenario.Difficulty,
		},
		Status:    string(result.Session.Status),
		TurnCount: result.Session.TurnCount,
		Messages:  toMessageResponses(result.Session.Messages),
		CreatedAt: formatTime(result.Session.CreatedAt),
		EndedAt:   formatOptionalTime(result.Session.EndedAt),
	}
}

// toMessageResponses 批量转换消息响应结构。
func toMessageResponses(messages []model.Message) []messageResponse {
	result := make([]messageResponse, 0, len(messages))
	for _, message := range messages {
		result = append(result, toMessageResponse(message))
	}

	return result
}

// toMessageResponse 将单条消息转换为 HTTP 响应结构。
func toMessageResponse(message model.Message) messageResponse {
	return messageResponse{
		ID:        message.ID,
		SessionID: message.SessionID,
		Role:      string(message.Role),
		Content:   message.Content,
		Stage:     message.Stage,
		CreatedAt: formatTime(message.CreatedAt),
	}
}

// formatOptionalTime 格式化可能为空的时间。
func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := formatTime(*value)
	return &formatted
}

// formatTime 使用统一 RFC3339 格式输出时间。
func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

// writeSessionError 将 Session 业务错误转换为统一 HTTP 响应。
func writeSessionError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidSessionRequest) {
		response.Error(c, http.StatusBadRequest, invalidSessionRequestCode, "训练请求无效")
		return
	}
	if errors.Is(err, service.ErrScenarioNotFound) {
		response.Error(c, http.StatusNotFound, scenarioNotFoundCode, "未找到训练场景")
		return
	}
	if errors.Is(err, service.ErrSessionNotFound) {
		response.Error(c, http.StatusNotFound, sessionNotFoundCode, "未找到训练")
		return
	}
	if errors.Is(err, service.ErrSessionAlreadyFinished) {
		response.Error(c, http.StatusConflict, sessionAlreadyFinishedCode, "训练已结束")
		return
	}
	if errors.Is(err, service.ErrStateStoreFailed) {
		response.Error(c, http.StatusServiceUnavailable, sessionStateStoreFailedCode, "训练短期状态写入失败")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "服务器内部错误")
}
