package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"speakmate/internal/response"
	"speakmate/internal/service"
)

const (
	invalidMessageRequestCode   = 3001
	messageContentRequiredCode  = 3002
	conversationAgentFailedCode = 3003
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
		response.Error(c, http.StatusBadRequest, invalidMessageRequestCode, "invalid message request")
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
		UserMessage: toMessageResponse(result.UserMessage),
		AIMessage:   toMessageResponse(result.AIMessage),
		Stage:       result.Stage,
		NextGoal:    result.NextGoal,
		TurnCount:   result.TurnCount,
	})
}

type sendMessageRequest struct {
	Content *string `json:"content"`
}

type sendMessageResponse struct {
	UserMessage messageResponse `json:"user_message"`
	AIMessage   messageResponse `json:"ai_message"`
	Stage       string          `json:"stage"`
	NextGoal    string          `json:"next_goal"`
	TurnCount   int             `json:"turn_count"`
}

func writeMessageError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidMessageRequest) {
		response.Error(c, http.StatusBadRequest, invalidMessageRequestCode, "invalid message request")
		return
	}
	if errors.Is(err, service.ErrMessageContentRequired) {
		response.Error(c, http.StatusBadRequest, messageContentRequiredCode, "message content is required")
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
	if errors.Is(err, service.ErrScenarioNotFound) {
		response.Error(c, http.StatusNotFound, scenarioNotFoundCode, "scenario not found")
		return
	}
	if errors.Is(err, service.ErrConversationAgentFailed) {
		response.Error(c, http.StatusBadGateway, conversationAgentFailedCode, "conversation agent failed")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
}
