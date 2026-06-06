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
	invalidFeedbackRequestCode = 4001
	correctionNotFoundCode     = 4002
	scoreNotFoundCode          = 4003
)

// FeedbackService 定义 Feedback Handler 依赖的业务能力。
type FeedbackService interface {
	GetMessageCorrection(messageID int) (model.CorrectionResult, error)
	ListSessionCorrections(sessionID int) ([]model.CorrectionResult, error)
	GetSessionCurrentScore(sessionID int) (model.ScoreResult, error)
}

// FeedbackHandler 负责处理纠错与评分查询 API。
type FeedbackHandler struct {
	service FeedbackService
}

// NewFeedbackHandler 创建 Feedback API Handler。
func NewFeedbackHandler(service FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{
		service: service,
	}
}

// GetMessageCorrection 查询单条用户消息纠错结果。
func (h *FeedbackHandler) GetMessageCorrection(c *gin.Context) {
	id, ok := parsePositiveMessageID(c)
	if !ok {
		return
	}

	correction, err := h.service.GetMessageCorrection(id)
	if err != nil {
		writeFeedbackError(c, err)
		return
	}

	response.Success(c, correction)
}

// ListSessionCorrections 查询某次训练的全部纠错结果。
func (h *FeedbackHandler) ListSessionCorrections(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}

	corrections, err := h.service.ListSessionCorrections(id)
	if err != nil {
		writeFeedbackError(c, err)
		return
	}

	response.Success(c, corrections)
}

// GetSessionScore 查询某次训练当前评分。
func (h *FeedbackHandler) GetSessionScore(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}

	score, err := h.service.GetSessionCurrentScore(id)
	if err != nil {
		writeFeedbackError(c, err)
		return
	}

	response.Success(c, score)
}

func parsePositiveMessageID(c *gin.Context) (int, bool) {
	rawID := c.Param("message_id")
	if rawID == "" {
		rawID = c.Param("id")
	}

	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, invalidFeedbackRequestCode, "invalid feedback request")
		return 0, false
	}

	return id, true
}

func writeFeedbackError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidFeedbackRequest) {
		response.Error(c, http.StatusBadRequest, invalidFeedbackRequestCode, "invalid feedback request")
		return
	}
	if errors.Is(err, service.ErrCorrectionNotFound) {
		response.Error(c, http.StatusNotFound, correctionNotFoundCode, "correction not found")
		return
	}
	if errors.Is(err, service.ErrScoreNotFound) {
		response.Error(c, http.StatusNotFound, scoreNotFoundCode, "score not found")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
}
