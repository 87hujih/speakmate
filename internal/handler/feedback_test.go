package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/service"
)

func TestFeedbackHandlerGetsSessionCurrentScore(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{
		score: model.ScoreResult{
			MessageID:  10,
			SessionID:  1,
			Fluency:    75,
			Grammar:    72,
			Expression: 80,
			Vocabulary: 76,
			Completion: 85,
			TotalScore: 78,
			Comment:    "用户能够表达核心意思，但存在时态和动词形式错误。",
		},
	}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:session_id/scores", handler.GetSessionScore)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/1/scores", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if feedbackService.lastSessionID != 1 {
		t.Fatalf("session id = %d, want 1", feedbackService.lastSessionID)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			MessageID  int    `json:"message_id"`
			SessionID  int    `json:"session_id"`
			Fluency    int    `json:"fluency"`
			Grammar    int    `json:"grammar"`
			Expression int    `json:"expression"`
			Vocabulary int    `json:"vocabulary"`
			Completion int    `json:"completion"`
			TotalScore int    `json:"total_score"`
			Comment    string `json:"comment"`
		} `json:"data"`
	}
	mustUnmarshalJSON(t, rec.Body.Bytes(), &body)

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want success", body.Message)
	}
	if body.Data.SessionID != 1 {
		t.Fatalf("session_id = %d, want 1", body.Data.SessionID)
	}
	if body.Data.MessageID != 10 {
		t.Fatalf("message_id = %d, want 10", body.Data.MessageID)
	}
	if body.Data.TotalScore != 78 {
		t.Fatalf("total_score = %d, want 78", body.Data.TotalScore)
	}
	if body.Data.Grammar != 72 {
		t.Fatalf("grammar = %d, want 72", body.Data.Grammar)
	}
	if body.Data.Comment == "" {
		t.Fatal("comment is empty")
	}
}

func TestFeedbackHandlerGetsMessageCorrection(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{
		correction: model.CorrectionResult{
			MessageID:     10,
			SessionID:     1,
			OriginalText:  "I am study computer science and I have did a project.",
			CorrectedText: "I am studying computer science, and I have done a project.",
			Errors: []model.CorrectionError{
				{
					Type:        model.CorrectionErrorTypeGrammar,
					Span:        "am study",
					Suggestion:  "am studying",
					Explanation: "be 动词后应接现在分词。",
				},
			},
			BetterExpressions: []string{
				"I major in computer science.",
				"I worked on a robotics project.",
			},
		},
	}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/messages/:message_id/corrections", handler.GetMessageCorrection)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/10/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if feedbackService.lastMessageID != 10 {
		t.Fatalf("message id = %d, want 10", feedbackService.lastMessageID)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			MessageID     int    `json:"message_id"`
			SessionID     int    `json:"session_id"`
			OriginalText  string `json:"original_text"`
			CorrectedText string `json:"corrected_text"`
			Errors        []struct {
				Type        string `json:"type"`
				Span        string `json:"span"`
				Suggestion  string `json:"suggestion"`
				Explanation string `json:"explanation"`
			} `json:"errors"`
			BetterExpressions []string `json:"better_expressions"`
		} `json:"data"`
	}
	mustUnmarshalJSON(t, rec.Body.Bytes(), &body)

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want success", body.Message)
	}
	if body.Data.MessageID != 10 {
		t.Fatalf("message_id = %d, want 10", body.Data.MessageID)
	}
	if body.Data.CorrectedText != "I am studying computer science, and I have done a project." {
		t.Fatalf("corrected_text = %q, want saved correction", body.Data.CorrectedText)
	}
	if len(body.Data.Errors) != 1 {
		t.Fatalf("errors length = %d, want 1", len(body.Data.Errors))
	}
	if body.Data.Errors[0].Type != "grammar" {
		t.Fatalf("error type = %q, want grammar", body.Data.Errors[0].Type)
	}
	if len(body.Data.BetterExpressions) != 2 {
		t.Fatalf("better_expressions length = %d, want 2", len(body.Data.BetterExpressions))
	}
}

func TestFeedbackHandlerListsSessionCorrections(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{
		corrections: []model.CorrectionResult{
			{
				MessageID:     10,
				SessionID:     1,
				OriginalText:  "I am study computer science.",
				CorrectedText: "I am studying computer science.",
				Errors: []model.CorrectionError{
					{
						Type:        model.CorrectionErrorTypeGrammar,
						Span:        "am study",
						Suggestion:  "am studying",
						Explanation: "be 动词后应接现在分词。",
					},
				},
				BetterExpressions: []string{"I major in computer science."},
			},
			{
				MessageID:         12,
				SessionID:         1,
				OriginalText:      "I worked on a project.",
				CorrectedText:     "I worked on a project.",
				Errors:            []model.CorrectionError{},
				BetterExpressions: []string{"I contributed to a robotics project."},
			},
		},
	}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:session_id/corrections", handler.ListSessionCorrections)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/1/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if feedbackService.lastSessionID != 1 {
		t.Fatalf("session id = %d, want 1", feedbackService.lastSessionID)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			MessageID     int    `json:"message_id"`
			SessionID     int    `json:"session_id"`
			OriginalText  string `json:"original_text"`
			CorrectedText string `json:"corrected_text"`
			Errors        []struct {
				Type        string `json:"type"`
				Span        string `json:"span"`
				Suggestion  string `json:"suggestion"`
				Explanation string `json:"explanation"`
			} `json:"errors"`
			BetterExpressions []string `json:"better_expressions"`
		} `json:"data"`
	}
	mustUnmarshalJSON(t, rec.Body.Bytes(), &body)

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want success", body.Message)
	}
	if len(body.Data) != 2 {
		t.Fatalf("corrections length = %d, want 2", len(body.Data))
	}
	if body.Data[0].MessageID != 10 {
		t.Fatalf("first message_id = %d, want 10", body.Data[0].MessageID)
	}
	if body.Data[1].MessageID != 12 {
		t.Fatalf("second message_id = %d, want 12", body.Data[1].MessageID)
	}
}

func TestFeedbackHandlerListSessionCorrectionsRejectsInvalidSessionID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:session_id/corrections", handler.ListSessionCorrections)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/abc/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertFeedbackErrorResponse(t, rec, http.StatusBadRequest, 2002, "invalid session id")
	if feedbackService.callCount != 0 {
		t.Fatalf("call count = %d, want 0", feedbackService.callCount)
	}
}

func TestFeedbackHandlerListSessionCorrectionsReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{err: service.ErrCorrectionNotFound}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:session_id/corrections", handler.ListSessionCorrections)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/999/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertFeedbackErrorResponse(t, rec, http.StatusNotFound, 4002, "correction not found")
}

func TestFeedbackHandlerGetMessageCorrectionRejectsInvalidMessageID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/messages/:message_id/corrections", handler.GetMessageCorrection)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/abc/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertFeedbackErrorResponse(t, rec, http.StatusBadRequest, 4001, "invalid feedback request")
	if feedbackService.callCount != 0 {
		t.Fatalf("call count = %d, want 0", feedbackService.callCount)
	}
}

func TestFeedbackHandlerGetMessageCorrectionReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{err: service.ErrCorrectionNotFound}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/messages/:message_id/corrections", handler.GetMessageCorrection)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/999/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertFeedbackErrorResponse(t, rec, http.StatusNotFound, 4002, "correction not found")
}

func TestFeedbackHandlerGetSessionScoreRejectsInvalidSessionID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:session_id/scores", handler.GetSessionScore)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/abc/scores", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertFeedbackErrorResponse(t, rec, http.StatusBadRequest, 2002, "invalid session id")
	if feedbackService.callCount != 0 {
		t.Fatalf("call count = %d, want 0", feedbackService.callCount)
	}
}

func TestFeedbackHandlerGetSessionScoreReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{err: service.ErrScoreNotFound}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:session_id/scores", handler.GetSessionScore)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/999/scores", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertFeedbackErrorResponse(t, rec, http.StatusNotFound, 4003, "score not found")
}

func TestFeedbackHandlerGetSessionScoreReturnsInternalError(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	feedbackService := &fakeFeedbackService{err: errors.New("store failed")}
	handler := NewFeedbackHandler(feedbackService)
	engine := gin.New()
	engine.GET("/api/v1/sessions/:session_id/scores", handler.GetSessionScore)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/1/scores", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertFeedbackErrorResponse(t, rec, http.StatusInternalServerError, http.StatusInternalServerError, "internal server error")
}

type fakeFeedbackService struct {
	correction    model.CorrectionResult
	corrections   []model.CorrectionResult
	score         model.ScoreResult
	err           error
	callCount     int
	lastMessageID int
	lastSessionID int
}

func (s *fakeFeedbackService) GetMessageCorrection(messageID int) (model.CorrectionResult, error) {
	s.callCount++
	s.lastMessageID = messageID
	if s.err != nil {
		return model.CorrectionResult{}, s.err
	}

	return s.correction, nil
}

func (s *fakeFeedbackService) ListSessionCorrections(sessionID int) ([]model.CorrectionResult, error) {
	s.callCount++
	s.lastSessionID = sessionID
	if s.err != nil {
		return nil, s.err
	}

	return s.corrections, nil
}

func (s *fakeFeedbackService) GetSessionCurrentScore(sessionID int) (model.ScoreResult, error) {
	s.callCount++
	s.lastSessionID = sessionID
	if s.err != nil {
		return model.ScoreResult{}, s.err
	}

	return s.score, nil
}

func assertFeedbackErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, httpStatus int, code int, message string) {
	t.Helper()

	if rec.Code != httpStatus {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, httpStatus, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	mustUnmarshalJSON(t, rec.Body.Bytes(), &body)

	if body.Code != code {
		t.Fatalf("code = %d, want %d", body.Code, code)
	}
	if body.Message != message {
		t.Fatalf("message = %q, want %q", body.Message, message)
	}
}

func mustUnmarshalJSON(t *testing.T, data []byte, v any) {
	t.Helper()

	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}
