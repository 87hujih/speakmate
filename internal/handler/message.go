package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/response"
	"speakmate/internal/service"
)

const (
	invalidMessageRequestCode  = 3001
	messageContentRequiredCode = 3002
)

// MessageService 定义 Handler 依赖的消息发送业务能力。
type MessageService interface {
	SendMessage(input service.SendMessageInput) (service.SendMessageResult, error)
}

// MessageHandler 负责处理 Message API 的 HTTP 请求。
type MessageHandler struct {
	service MessageService
}

// NewMessageHandler 创建 Message API Handler。
func NewMessageHandler(service MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

// Send 向 running Session 发送用户文本消息并返回 Mock AI 回复。
func (h *MessageHandler) Send(c *gin.Context) {
	sessionID, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == nil {
		response.Error(c, http.StatusBadRequest, invalidMessageRequestCode, "invalid message request")
		return
	}

	result, err := h.service.SendMessage(service.SendMessageInput{
		SessionID: sessionID,
		Content:   *req.Content,
	})
	if err != nil {
		writeMessageError(c, err)
		return
	}

	response.Success(c, sendMessageResponse{
		UserMessage: toMessageResponse(result.UserMessage),
		AIMessage:   toMessageResponse(result.AIMessage),
		Stage:       result.Stage,
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
	TurnCount   int             `json:"turn_count"`
}

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

func writeMessageError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidMessageRequest) {
		response.Error(c, http.StatusBadRequest, invalidMessageRequestCode, "invalid message request")
		return
	}
	if errors.Is(err, service.ErrMessageContentRequired) {
		response.Error(c, http.StatusBadRequest, messageContentRequiredCode, "message content is required")
		return
	}

	writeSessionError(c, err)
}
