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

// 当前模块使用的业务错误码和事件常量。
const (
	invalidReportRequestCode  = 5001
	sessionNotFinishedCode    = 5002
	reportNotFoundCode        = 5003
	reportFeedbackMissingCode = 5004
	summaryAgentFailedCode    = 5005
)

// ReportService 定义 Report Handler 依赖的业务能力。
type ReportService interface {
	GenerateReport(sessionID int) (model.Report, error)
	GetReport(sessionID int) (model.Report, error)
}

// ReportHandler 负责处理课后报告 API。
type ReportHandler struct {
	service ReportService
}

// NewReportHandler 创建 Report API Handler。
func NewReportHandler(service ReportService) *ReportHandler {
	return &ReportHandler{
		service: service,
	}
}

// Generate 生成并保存课后报告。
func (h *ReportHandler) Generate(c *gin.Context) {
	id, ok := parsePositiveReportSessionID(c)
	if !ok {
		return
	}

	report, err := h.service.GenerateReport(id)
	if err != nil {
		writeReportError(c, err)
		return
	}

	response.Success(c, report)
}

// Get 查询已生成的课后报告。
func (h *ReportHandler) Get(c *gin.Context) {
	id, ok := parsePositiveReportSessionID(c)
	if !ok {
		return
	}

	report, err := h.service.GetReport(id)
	if err != nil {
		writeReportError(c, err)
		return
	}

	response.Success(c, report)
}

// parsePositiveReportSessionID 从路径参数解析报告 Session ID。
func parsePositiveReportSessionID(c *gin.Context) (int, bool) {
	rawID := c.Param("id")
	if rawID == "" {
		rawID = c.Param("session_id")
	}

	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, invalidReportRequestCode, "报告请求无效")
		return 0, false
	}

	return id, true
}

// writeReportError 将报告业务错误转换为统一 HTTP 响应。
func writeReportError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidReportRequest) {
		response.Error(c, http.StatusBadRequest, invalidReportRequestCode, "报告请求无效")
		return
	}
	if errors.Is(err, service.ErrSessionNotFound) {
		response.Error(c, http.StatusNotFound, sessionNotFoundCode, "未找到训练")
		return
	}
	if errors.Is(err, service.ErrSessionNotFinished) {
		response.Error(c, http.StatusConflict, sessionNotFinishedCode, "训练尚未结束")
		return
	}
	if errors.Is(err, service.ErrReportNotFound) {
		response.Error(c, http.StatusNotFound, reportNotFoundCode, "未找到课后报告")
		return
	}
	if errors.Is(err, service.ErrReportFeedbackMissing) {
		response.Error(c, http.StatusConflict, reportFeedbackMissingCode, "报告缺少反馈数据")
		return
	}
	if errors.Is(err, service.ErrSummaryAgentFailed) {
		response.Error(c, http.StatusBadGateway, summaryAgentFailedCode, "报告摘要生成失败")
		return
	}
	if errors.Is(err, service.ErrEventPublishFailed) {
		response.Error(c, http.StatusServiceUnavailable, streamEventPublishFailedCode, "实时事件发布失败")
		return
	}

	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "服务器内部错误")
}
