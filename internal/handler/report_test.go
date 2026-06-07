package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/service"
)

func TestReportHandlerGeneratesReport(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	reportService := &fakeReportService{report: sampleHandlerReport(7)}
	reportHandler := NewReportHandler(reportService)
	engine := gin.New()
	engine.POST("/api/v1/sessions/:id/report", reportHandler.Generate)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/7/report", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if reportService.generatedSessionID != 7 {
		t.Fatalf("generated session id = %d, want 7", reportService.generatedSessionID)
	}

	var body struct {
		Code int `json:"code"`
		Data struct {
			SessionID       int      `json:"session_id"`
			DurationSeconds int      `json:"duration_seconds"`
			TotalScore      int      `json:"total_score"`
			Summary         string   `json:"summary"`
			FrequentErrors  []string `json:"frequent_errors"`
		} `json:"data"`
	}
	mustUnmarshalReportJSON(t, rec.Body.Bytes(), &body)

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Data.SessionID != 7 {
		t.Fatalf("session_id = %d, want 7", body.Data.SessionID)
	}
	if body.Data.DurationSeconds != 180 {
		t.Fatalf("duration_seconds = %d, want 180", body.Data.DurationSeconds)
	}
	if body.Data.TotalScore != 77 {
		t.Fatalf("total_score = %d, want 77", body.Data.TotalScore)
	}
	if body.Data.Summary == "" {
		t.Fatal("summary is empty")
	}
	if len(body.Data.FrequentErrors) != 1 {
		t.Fatalf("frequent_errors length = %d, want 1", len(body.Data.FrequentErrors))
	}
}

func TestReportHandlerGetsReport(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	reportService := &fakeReportService{report: sampleHandlerReport(7)}
	reportHandler := NewReportHandler(reportService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:id/report", reportHandler.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/7/report", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if reportService.gotSessionID != 7 {
		t.Fatalf("got session id = %d, want 7", reportService.gotSessionID)
	}
}

func TestReportHandlerRejectsInvalidSessionID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	reportService := &fakeReportService{}
	reportHandler := NewReportHandler(reportService)
	engine := gin.New()
	engine.POST("/api/v1/sessions/:id/report", reportHandler.Generate)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/abc/report", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertReportErrorResponse(t, rec, http.StatusBadRequest, 5001, "invalid report request")
	if reportService.callCount != 0 {
		t.Fatalf("call count = %d, want 0", reportService.callCount)
	}
}

func TestReportHandlerMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
		wantMsg    string
	}{
		{name: "session not found", err: service.ErrSessionNotFound, wantStatus: http.StatusNotFound, wantCode: 2003, wantMsg: "session not found"},
		{name: "session not finished", err: service.ErrSessionNotFinished, wantStatus: http.StatusConflict, wantCode: 5002, wantMsg: "session not finished"},
		{name: "report not found", err: service.ErrReportNotFound, wantStatus: http.StatusNotFound, wantCode: 5003, wantMsg: "report not found"},
		{name: "feedback missing", err: service.ErrReportFeedbackMissing, wantStatus: http.StatusConflict, wantCode: 5004, wantMsg: "report feedback missing"},
		{name: "summary failed", err: service.ErrSummaryAgentFailed, wantStatus: http.StatusBadGateway, wantCode: 5005, wantMsg: "summary agent failed"},
		{name: "internal error", err: errors.New("store failed"), wantStatus: http.StatusInternalServerError, wantCode: http.StatusInternalServerError, wantMsg: "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.ReleaseMode)
			reportService := &fakeReportService{err: tt.err}
			reportHandler := NewReportHandler(reportService)
			engine := gin.New()
			engine.POST("/api/v1/sessions/:id/report", reportHandler.Generate)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/7/report", nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assertReportErrorResponse(t, rec, tt.wantStatus, tt.wantCode, tt.wantMsg)
		})
	}
}

type fakeReportService struct {
	report             model.Report
	err                error
	callCount          int
	generatedSessionID int
	gotSessionID       int
}

func (s *fakeReportService) GenerateReport(sessionID int) (model.Report, error) {
	s.callCount++
	s.generatedSessionID = sessionID
	if s.err != nil {
		return model.Report{}, s.err
	}

	return s.report, nil
}

func (s *fakeReportService) GetReport(sessionID int) (model.Report, error) {
	s.callCount++
	s.gotSessionID = sessionID
	if s.err != nil {
		return model.Report{}, s.err
	}

	return s.report, nil
}

func sampleHandlerReport(sessionID int) model.Report {
	return model.Report{
		SessionID: sessionID,
		Scenario: model.ReportScenario{
			ID:         1,
			Code:       "interview",
			Name:       "英语面试",
			Difficulty: "medium",
		},
		DurationSeconds: 180,
		TurnCount:       1,
		TotalScore:      77,
		Scores:          model.ScoreResult{SessionID: sessionID, TotalScore: 77},
		Summary:         "本次训练能够说明项目背景，但需要加强动词形式。",
		MajorProblems:   []string{"动词形式不稳定"},
		FrequentErrors:  []string{"am study -> am studying"},
		BetterExpressions: []string{
			"I major in computer science.",
		},
		NextPracticePlan: []string{"用 STAR 结构重写项目经历回答。"},
		CreatedAt:        time.Date(2026, 6, 7, 3, 5, 0, 0, time.UTC),
	}
}

func assertReportErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, httpStatus int, code int, message string) {
	t.Helper()

	if rec.Code != httpStatus {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, httpStatus, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	mustUnmarshalReportJSON(t, rec.Body.Bytes(), &body)

	if body.Code != code {
		t.Fatalf("code = %d, want %d", body.Code, code)
	}
	if body.Message != message {
		t.Fatalf("message = %q, want %q", body.Message, message)
	}
}

func mustUnmarshalReportJSON(t *testing.T, data []byte, v any) {
	t.Helper()

	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}
