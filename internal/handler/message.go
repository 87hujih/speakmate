package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"speakmate/internal/response"
	"speakmate/internal/service"
)

// 当前模块使用的业务错误码和事件常量。
const (
	invalidMessageRequestCode    = 3001
	messageContentRequiredCode   = 3002
	conversationAgentFailedCode  = 3003
	feedbackAgentFailedCode      = 3004
	sessionStateStoreFailedCode  = 8001
	streamEventPublishFailedCode = 8002
)

// MessageService 定义消息 Handler 依赖的业务能力。
type MessageService interface {
	SendMessage(input service.SendMessageInput) (service.SendMessageResult, error)
}

// MessageHandler 负责处理消息发送 API。
type MessageHandler struct {
	service MessageService
}

// NewMessageHandler 创建 Message API Handler。
func NewMessageHandler(service MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

// Send 向 running Session 发送一条用户消息，并返回 AI 回复。
func (h *MessageHandler) Send(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == nil {
		response.Error(c, http.StatusBadRequest, invalidMessageRequestCode, "消息请求无效")
		return
	}

	result, err := h.service.SendMessage(service.SendMessageInput{
		SessionID: id,
		Content:   *req.Content,
		Context:   c.Request.Context(),
	})
	if err != nil {
		writeMessageError(c, err)
		return
	}

	response.Success(c, sendMessageResponse{
		UserMessage:       toMessageResponse(result.UserMessage),
		AIMessage:         toMessageResponse(result.AIMessage),
		Stage:             result.Stage,
		NextGoal:          result.NextGoal,
		TurnCount:         result.TurnCount,
		CorrectionSummary: toCorrectionSummaryResponse(result.CorrectionSummary),
		ScoreSummary:      toScoreSummaryResponse(result.ScoreSummary),
	})
}

// sendMessageRequest 是发送文本消息接口请求结构。
type sendMessageRequest struct {
	Content *string `json:"content"`
}

// sendMessageResponse 是发送文本消息接口返回结构。
type sendMessageResponse struct {
	UserMessage       messageResponse           `json:"user_message"`
	AIMessage         messageResponse           `json:"ai_message"`
	Stage             string                    `json:"stage"`
	NextGoal          string                    `json:"next_goal"`
	TurnCount         int                       `json:"turn_count"`
	CorrectionSummary correctionSummaryResponse `json:"correction_summary"`
	ScoreSummary      scoreSummaryResponse      `json:"score_summary"`
}

// correctionSummaryResponse 是消息响应中的纠错摘要结构。
type correctionSummaryResponse struct {
	HasErrors  bool `json:"has_errors"`
	ErrorCount int  `json:"error_count"`
}

// scoreSummaryResponse 是消息响应中的评分摘要结构。
type scoreSummaryResponse struct {
	TotalScore int `json:"total_score"`
	Grammar    int `json:"grammar"`
	Expression int `json:"expression"`
}

// toCorrectionSummaryResponse 将纠错摘要转换为消息响应结构。
func toCorrectionSummaryResponse(summary service.CorrectionSummary) correctionSummaryResponse {
	return correctionSummaryResponse{
		HasErrors:  summary.HasErrors,
		ErrorCount: summary.ErrorCount,
	}
}

// toScoreSummaryResponse 将评分摘要转换为消息响应结构。
func toScoreSummaryResponse(summary service.ScoreSummary) scoreSummaryResponse {
	return scoreSummaryResponse{
		TotalScore: summary.TotalScore,
		Grammar:    summary.Grammar,
		Expression: summary.Expression,
	}
}

// writeMessageError 将消息发送错误转换为统一 HTTP 响应。
func writeMessageError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidMessageRequest) {
		response.Error(c, http.StatusBadRequest, invalidMessageRequestCode, "消息请求无效")
		return
	}
	if errors.Is(err, service.ErrMessageContentRequired) {
		response.Error(c, http.StatusBadRequest, messageContentRequiredCode, "消息内容不能为空")
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
	if errors.Is(err, service.ErrScenarioNotFound) {
		response.Error(c, http.StatusNotFound, scenarioNotFoundCode, "未找到训练场景")
		return
	}
	if errors.Is(err, service.ErrConversationAgentFailed) {
		response.Error(c, http.StatusBadGateway, conversationAgentFailedCode, "对话 AI 回复失败")
		return
	}
	if errors.Is(err, service.ErrFeedbackAgentFailed) {
		response.Error(c, http.StatusBadGateway, feedbackAgentFailedCode, "反馈 AI 生成失败")
		return
	}
	if errors.Is(err, service.ErrStateStoreFailed) {
		response.Error(c, http.StatusServiceUnavailable, sessionStateStoreFailedCode, "训练短期状态写入失败")
		return
	}
	if errors.Is(err, service.ErrEventPublishFailed) {
		response.Error(c, http.StatusServiceUnavailable, streamEventPublishFailedCode, "实时事件发布失败")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "服务器内部错误")
}
