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
		response.Error(c, http.StatusBadRequest, invalidSessionRequestCode, "invalid session request")
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

type createSessionRequest struct {
	ScenarioID int `json:"scenario_id"`
	UserID     int `json:"user_id"`
}

type createSessionResponse struct {
	SessionID      int    `json:"session_id"`
	SessionNo      string `json:"session_no"`
	ScenarioID     int    `json:"scenario_id"`
	Status         string `json:"status"`
	OpeningMessage string `json:"opening_message"`
}

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

type messageResponse struct {
	ID        int    `json:"id"`
	SessionID int    `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type finishSessionResponse struct {
	SessionID int    `json:"session_id"`
	Status    string `json:"status"`
	TurnCount int    `json:"turn_count"`
	EndedAt   string `json:"ended_at"`
}

func parsePositiveSessionID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, invalidSessionIDCode, "invalid session id")
		return 0, false
	}

	return id, true
}

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

func toMessageResponses(messages []model.Message) []messageResponse {
	result := make([]messageResponse, 0, len(messages))
	for _, message := range messages {
		result = append(result, messageResponse{
			ID:        message.ID,
			SessionID: message.SessionID,
			Role:      string(message.Role),
			Content:   message.Content,
			CreatedAt: formatTime(message.CreatedAt),
		})
	}

	return result
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := formatTime(*value)
	return &formatted
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func writeSessionError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidSessionRequest) {
		response.Error(c, http.StatusBadRequest, invalidSessionRequestCode, "invalid session request")
		return
	}
	if errors.Is(err, service.ErrScenarioNotFound) {
		response.Error(c, http.StatusNotFound, scenarioNotFoundCode, "scenario not found")
		return
	}
	if errors.Is(err, service.ErrSessionNotFound) {
		response.Error(c, http.StatusNotFound, sessionNotFoundCode, "session not found")
		return
	}
	if errors.Is(err, service.ErrSessionAlreadyFinished) {
		response.Error(c, http.StatusConflict, sessionAlreadyFinishedCode, "session already finished")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
}
