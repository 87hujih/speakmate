package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"speakmate/internal/response"
	"speakmate/internal/service"
)

// 当前模块使用的业务错误码和事件常量。
const (
	invalidAudioRequestCode        = 7001
	audioFileRequiredCode          = 7002
	audioFileTooLargeCode          = 7003
	audioFileTypeUnsupportedCode   = 7004
	asrClientFailedCode            = 7005
	audioTranscriptRequiredCode    = 7006
	audioMultipartMemoryLimitBytes = 512 * 1024
	audioReadLimitBytes            = service.MaxAudioUploadBytes + 1
)

// AudioService 定义音频上传 Handler 依赖的业务能力。
type AudioService interface {
	UploadAudio(input service.UploadAudioInput) (service.UploadAudioResult, error)
}

// AudioHandler 负责处理音频上传 API。
type AudioHandler struct {
	service AudioService
}

// NewAudioHandler 创建 Audio API Handler。
func NewAudioHandler(service AudioService) *AudioHandler {
	return &AudioHandler{
		service: service,
	}
}

// Upload 接收单段音频文件，转写后复用消息发送流程。
func (h *AudioHandler) Upload(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}

	if err := c.Request.ParseMultipartForm(audioMultipartMemoryLimitBytes); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		response.Error(c, http.StatusBadRequest, invalidAudioRequestCode, "音频请求无效")
		return
	}

	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		response.Error(c, http.StatusBadRequest, audioFileRequiredCode, "请上传音频文件")
		return
	}
	defer file.Close()

	audio, err := io.ReadAll(io.LimitReader(file, audioReadLimitBytes))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "服务器内部错误")
		return
	}

	contentType := header.Header.Get("Content-Type")
	result, err := h.service.UploadAudio(service.UploadAudioInput{
		SessionID:   id,
		Filename:    header.Filename,
		ContentType: contentType,
		Audio:       audio,
		Context:     c.Request.Context(),
	})
	if err != nil {
		writeAudioError(c, err)
		return
	}

	response.Success(c, uploadAudioResponse{
		Transcript:        result.Transcript,
		UserMessage:       toMessageResponse(result.UserMessage),
		AIMessage:         toMessageResponse(result.AIMessage),
		Stage:             result.Stage,
		NextGoal:          result.NextGoal,
		TurnCount:         result.TurnCount,
		CorrectionSummary: toCorrectionSummaryResponse(result.CorrectionSummary),
		ScoreSummary:      toScoreSummaryResponse(result.ScoreSummary),
	})
}

// uploadAudioResponse 是单段音频上传接口返回结构。
type uploadAudioResponse struct {
	Transcript        string                    `json:"transcript"`
	UserMessage       messageResponse           `json:"user_message"`
	AIMessage         messageResponse           `json:"ai_message"`
	Stage             string                    `json:"stage"`
	NextGoal          string                    `json:"next_goal"`
	TurnCount         int                       `json:"turn_count"`
	CorrectionSummary correctionSummaryResponse `json:"correction_summary"`
	ScoreSummary      scoreSummaryResponse      `json:"score_summary"`
}

// writeAudioError 将音频业务错误转换为统一 HTTP 响应。
func writeAudioError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidAudioRequest) {
		response.Error(c, http.StatusBadRequest, invalidAudioRequestCode, "音频请求无效")
		return
	}
	if errors.Is(err, service.ErrAudioFileRequired) {
		response.Error(c, http.StatusBadRequest, audioFileRequiredCode, "请上传音频文件")
		return
	}
	if errors.Is(err, service.ErrAudioFileTooLarge) {
		response.Error(c, http.StatusRequestEntityTooLarge, audioFileTooLargeCode, "音频文件过大")
		return
	}
	if errors.Is(err, service.ErrAudioFileTypeUnsupported) {
		response.Error(c, http.StatusBadRequest, audioFileTypeUnsupportedCode, "不支持该音频格式")
		return
	}
	if errors.Is(err, service.ErrASRClientFailed) {
		response.Error(c, http.StatusBadGateway, asrClientFailedCode, "语音识别服务调用失败")
		return
	}
	if errors.Is(err, service.ErrAudioTranscriptRequired) {
		response.Error(c, http.StatusBadRequest, audioTranscriptRequiredCode, "语音识别未返回有效文本")
		return
	}

	writeMessageError(c, err)
}
