package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("LLM_USE_MOCK", "true")
	os.Exit(m.Run())
}

// TestHealthEndpointReturnsOK 验证健康检查接口保持统一成功响应结构。
func TestHealthEndpointReturnsOK(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want %q", body.Message, "success")
	}
	if body.Data.Status != "ok" {
		t.Fatalf("data.status = %q, want %q", body.Data.Status, "ok")
	}
}

// TestScenarioListReturnsBuiltInScenarios 验证场景列表接口返回 3 个内置场景。
func TestScenarioListReturnsBuiltInScenarios(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			ID          int    `json:"id"`
			Code        string `json:"code"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Difficulty  string `json:"difficulty"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want %q", body.Message, "success")
	}
	if len(body.Data) != 3 {
		t.Fatalf("scenario count = %d, want 3", len(body.Data))
	}
	if body.Data[0].Code != "interview" {
		t.Fatalf("first scenario code = %q, want %q", body.Data[0].Code, "interview")
	}
}

// TestScenarioDetailReturnsInterviewScenario 验证场景详情接口返回英语面试完整信息。
func TestScenarioDetailReturnsInterviewScenario(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios/1", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID             int    `json:"id"`
			Code           string `json:"code"`
			Name           string `json:"name"`
			Description    string `json:"description"`
			Difficulty     string `json:"difficulty"`
			AIRole         string `json:"ai_role"`
			UserGoal       string `json:"user_goal"`
			OpeningMessage string `json:"opening_message"`
			Stages         []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"stages"`
			Rubric []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"rubric"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want %q", body.Message, "success")
	}
	if body.Data.Code != "interview" {
		t.Fatalf("scenario code = %q, want %q", body.Data.Code, "interview")
	}
	if body.Data.AIRole != "技术面试官" {
		t.Fatalf("ai_role = %q, want %q", body.Data.AIRole, "技术面试官")
	}
	if body.Data.OpeningMessage == "" {
		t.Fatal("opening_message is empty")
	}
	if len(body.Data.Stages) == 0 {
		t.Fatal("stages is empty")
	}
	if len(body.Data.Rubric) == 0 {
		t.Fatal("rubric is empty")
	}
}

// TestScenarioDetailReturnsNotFoundError 验证不存在的场景返回统一 404 错误响应。
func TestScenarioDetailReturnsNotFoundError(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios/999", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 1001 {
		t.Fatalf("code = %d, want 1001", body.Code)
	}
	if body.Message != "scenario not found" {
		t.Fatalf("message = %q, want %q", body.Message, "scenario not found")
	}
}

// TestScenarioDetailReturnsInvalidIDError 验证非法场景 ID 返回统一 400 错误响应。
func TestScenarioDetailReturnsInvalidIDError(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios/abc", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 1002 {
		t.Fatalf("code = %d, want 1002", body.Code)
	}
	if body.Message != "invalid scenario id" {
		t.Fatalf("message = %q, want %q", body.Message, "invalid scenario id")
	}
}

// TestSessionCreateReturnsRunningSession 验证创建训练 Session 后直接进入 running 状态。
func TestSessionCreateReturnsRunningSession(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(`{"scenario_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			SessionID      int    `json:"session_id"`
			SessionNo      string `json:"session_no"`
			ScenarioID     int    `json:"scenario_id"`
			Status         string `json:"status"`
			OpeningMessage string `json:"opening_message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want %q", body.Message, "success")
	}
	if body.Data.SessionID <= 0 {
		t.Fatalf("session_id = %d, want positive", body.Data.SessionID)
	}
	if !strings.HasPrefix(body.Data.SessionNo, "S") {
		t.Fatalf("session_no = %q, want S prefix", body.Data.SessionNo)
	}
	if body.Data.ScenarioID != 1 {
		t.Fatalf("scenario_id = %d, want 1", body.Data.ScenarioID)
	}
	if body.Data.Status != "running" {
		t.Fatalf("status = %q, want %q", body.Data.Status, "running")
	}
	if body.Data.OpeningMessage == "" {
		t.Fatal("opening_message is empty")
	}
}

// TestSessionCreateRejectsInvalidRequest 验证创建 Session 时缺失或非法 scenario_id 会返回统一 400。
func TestSessionCreateRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing scenario_id", body: `{}`},
		{name: "zero scenario_id", body: `{"scenario_id":0}`},
		{name: "invalid json", body: `{`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assertErrorResponse(t, rec, http.StatusBadRequest, 2001, "invalid session request")
		})
	}
}

// TestSessionCreateReturnsScenarioNotFound 验证创建 Session 时不存在的场景复用场景不存在错误语义。
func TestSessionCreateReturnsScenarioNotFound(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(`{"scenario_id":999}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, 1001, "scenario not found")
}

// TestSessionGetReturnsCreatedSession 验证查询 Session 会返回场景摘要、轮次和空消息列表。
func TestSessionGetReturnsCreatedSession(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID), nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			SessionID int    `json:"session_id"`
			SessionNo string `json:"session_no"`
			Scenario  struct {
				ID          int    `json:"id"`
				Code        string `json:"code"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Difficulty  string `json:"difficulty"`
			} `json:"scenario"`
			Status    string        `json:"status"`
			TurnCount int           `json:"turn_count"`
			Messages  []interface{} `json:"messages"`
			CreatedAt string        `json:"created_at"`
			EndedAt   *string       `json:"ended_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want %q", body.Message, "success")
	}
	if body.Data.SessionID != sessionID {
		t.Fatalf("session_id = %d, want %d", body.Data.SessionID, sessionID)
	}
	if body.Data.SessionNo == "" {
		t.Fatal("session_no is empty")
	}
	if body.Data.Scenario.Code != "interview" {
		t.Fatalf("scenario.code = %q, want %q", body.Data.Scenario.Code, "interview")
	}
	if body.Data.Status != "running" {
		t.Fatalf("status = %q, want %q", body.Data.Status, "running")
	}
	if body.Data.TurnCount != 0 {
		t.Fatalf("turn_count = %d, want 0", body.Data.TurnCount)
	}
	if len(body.Data.Messages) != 0 {
		t.Fatalf("messages length = %d, want 0", len(body.Data.Messages))
	}
	if _, err := time.Parse(time.RFC3339, body.Data.CreatedAt); err != nil {
		t.Fatalf("created_at is not RFC3339: %v", err)
	}
	if body.Data.EndedAt != nil {
		t.Fatalf("ended_at = %v, want nil", *body.Data.EndedAt)
	}
}

// TestSessionGetReturnsNotFound 验证查询不存在的 Session 会返回统一 404。
func TestSessionGetReturnsNotFound(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/999", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, 2003, "session not found")
}

// TestSessionInvalidIDReturnsBadRequest 验证 Session 路径 ID 非正整数时返回统一 400。
func TestSessionInvalidIDReturnsBadRequest(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "get non-number", method: http.MethodGet, path: "/api/v1/sessions/abc"},
		{name: "get zero", method: http.MethodGet, path: "/api/v1/sessions/0"},
		{name: "finish non-number", method: http.MethodPost, path: "/api/v1/sessions/abc/finish"},
		{name: "finish zero", method: http.MethodPost, path: "/api/v1/sessions/0/finish"},
		{name: "message non-number", method: http.MethodPost, path: "/api/v1/sessions/abc/messages"},
		{name: "message zero", method: http.MethodPost, path: "/api/v1/sessions/0/messages"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assertErrorResponse(t, rec, http.StatusBadRequest, 2002, "invalid session id")
		})
	}
}

// TestSessionFinishMovesRunningSessionToFinished 验证结束 Session 后状态和结束时间会被保存。
func TestSessionFinishMovesRunningSessionToFinished(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/finish", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			SessionID int    `json:"session_id"`
			Status    string `json:"status"`
			TurnCount int    `json:"turn_count"`
			EndedAt   string `json:"ended_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want %q", body.Message, "success")
	}
	if body.Data.SessionID != sessionID {
		t.Fatalf("session_id = %d, want %d", body.Data.SessionID, sessionID)
	}
	if body.Data.Status != "finished" {
		t.Fatalf("status = %q, want %q", body.Data.Status, "finished")
	}
	if body.Data.TurnCount != 0 {
		t.Fatalf("turn_count = %d, want 0", body.Data.TurnCount)
	}
	if _, err := time.Parse(time.RFC3339, body.Data.EndedAt); err != nil {
		t.Fatalf("ended_at is not RFC3339: %v", err)
	}
}

// TestSessionFinishTwiceReturnsConflict 验证重复结束不会重写状态，而是返回业务冲突错误。
func TestSessionFinishTwiceReturnsConflict(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/finish", nil)
	firstRec := httptest.NewRecorder()
	engine.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusOK {
		t.Fatalf("first finish status code = %d, want %d; body = %s", firstRec.Code, http.StatusOK, firstRec.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/finish", nil)
	secondRec := httptest.NewRecorder()
	engine.ServeHTTP(secondRec, secondReq)

	assertErrorResponse(t, secondRec, http.StatusConflict, 2004, "session already finished")
}

// TestMessageSendCreatesMockReplyAndSessionHistory 验证发送文本消息后会保存用户消息、AI 回复和轮次。
func TestMessageSendCreatesMockReplyAndSessionHistory(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/sessions/"+strconv.Itoa(sessionID)+"/messages",
		bytes.NewBufferString(`{"content":" I worked on a robot control project. "}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			UserMessage messagePayload `json:"user_message"`
			AIMessage   messagePayload `json:"ai_message"`
			Stage       string         `json:"stage"`
			NextGoal    string         `json:"next_goal"`
			TurnCount   int            `json:"turn_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != 0 {
		t.Fatalf("code = %d, want 0", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("message = %q, want %q", body.Message, "success")
	}
	if body.Data.UserMessage.Role != "user" {
		t.Fatalf("user role = %q, want user", body.Data.UserMessage.Role)
	}
	if body.Data.UserMessage.Content != "I worked on a robot control project." {
		t.Fatalf("user content = %q, want trimmed content", body.Data.UserMessage.Content)
	}
	if body.Data.AIMessage.Role != "ai" {
		t.Fatalf("ai role = %q, want ai", body.Data.AIMessage.Role)
	}
	if !strings.Contains(strings.ToLower(body.Data.AIMessage.Content), "project") {
		t.Fatalf("ai content = %q, want interview project follow-up", body.Data.AIMessage.Content)
	}
	if body.Data.Stage == "" {
		t.Fatal("stage is empty")
	}
	if body.Data.NextGoal == "" {
		t.Fatal("next_goal is empty")
	}
	if body.Data.TurnCount != 1 {
		t.Fatalf("turn_count = %d, want 1", body.Data.TurnCount)
	}
	assertRFC3339(t, body.Data.UserMessage.CreatedAt)
	assertRFC3339(t, body.Data.AIMessage.CreatedAt)

	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID), nil)
	detailRec := httptest.NewRecorder()

	engine.ServeHTTP(detailRec, detailReq)

	if detailRec.Code != http.StatusOK {
		t.Fatalf("detail status code = %d, want %d; body = %s", detailRec.Code, http.StatusOK, detailRec.Body.String())
	}

	var detailBody struct {
		Data struct {
			TurnCount int              `json:"turn_count"`
			Messages  []messagePayload `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detailBody); err != nil {
		t.Fatalf("detail response is not valid JSON: %v", err)
	}
	if detailBody.Data.TurnCount != 1 {
		t.Fatalf("detail turn_count = %d, want 1", detailBody.Data.TurnCount)
	}
	if len(detailBody.Data.Messages) != 2 {
		t.Fatalf("detail messages length = %d, want 2", len(detailBody.Data.Messages))
	}
	if detailBody.Data.Messages[0].Role != "user" || detailBody.Data.Messages[1].Role != "ai" {
		t.Fatalf("detail message roles = %q/%q, want user/ai", detailBody.Data.Messages[0].Role, detailBody.Data.Messages[1].Role)
	}
}

// TestMessageSendGeneratesFeedbackAndFeedbackRoutesReturnIt 验证发送消息会同步生成纠错、评分，并能被反馈查询接口读到。
func TestMessageSendGeneratesFeedbackAndFeedbackRoutesReturnIt(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	sent := postMessage(t, engine, sessionID, `{"content":"I am study computer science and I have did a project."}`)

	if !sent.Data.CorrectionSummary.HasErrors {
		t.Fatal("correction_summary.has_errors = false, want true")
	}
	if sent.Data.CorrectionSummary.ErrorCount != 2 {
		t.Fatalf("correction_summary.error_count = %d, want 2", sent.Data.CorrectionSummary.ErrorCount)
	}
	if sent.Data.ScoreSummary.TotalScore != 77 {
		t.Fatalf("score_summary.total_score = %d, want 77", sent.Data.ScoreSummary.TotalScore)
	}
	if sent.Data.ScoreSummary.Grammar != 72 {
		t.Fatalf("score_summary.grammar = %d, want 72", sent.Data.ScoreSummary.Grammar)
	}
	if sent.Data.ScoreSummary.Expression != 80 {
		t.Fatalf("score_summary.expression = %d, want 80", sent.Data.ScoreSummary.Expression)
	}

	messageID := sent.Data.UserMessage.ID
	correctionReq := httptest.NewRequest(http.MethodGet, "/api/v1/messages/"+strconv.Itoa(messageID)+"/corrections", nil)
	correctionRec := httptest.NewRecorder()

	engine.ServeHTTP(correctionRec, correctionReq)

	if correctionRec.Code != http.StatusOK {
		t.Fatalf("correction status code = %d, want %d; body = %s", correctionRec.Code, http.StatusOK, correctionRec.Body.String())
	}
	var correctionBody struct {
		Code int `json:"code"`
		Data struct {
			MessageID     int    `json:"message_id"`
			SessionID     int    `json:"session_id"`
			OriginalText  string `json:"original_text"`
			CorrectedText string `json:"corrected_text"`
			Errors        []struct {
				Type       string `json:"type"`
				Span       string `json:"span"`
				Suggestion string `json:"suggestion"`
			} `json:"errors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(correctionRec.Body.Bytes(), &correctionBody); err != nil {
		t.Fatalf("correction response is not valid JSON: %v", err)
	}
	if correctionBody.Code != 0 {
		t.Fatalf("correction code = %d, want 0", correctionBody.Code)
	}
	if correctionBody.Data.MessageID != messageID {
		t.Fatalf("correction message_id = %d, want %d", correctionBody.Data.MessageID, messageID)
	}
	if correctionBody.Data.SessionID != sessionID {
		t.Fatalf("correction session_id = %d, want %d", correctionBody.Data.SessionID, sessionID)
	}
	if correctionBody.Data.CorrectedText != "I am studying computer science, and I have done a project." {
		t.Fatalf("corrected_text = %q, want mock corrected text", correctionBody.Data.CorrectedText)
	}
	if len(correctionBody.Data.Errors) != 2 {
		t.Fatalf("correction errors length = %d, want 2", len(correctionBody.Data.Errors))
	}

	sessionCorrectionsReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/corrections", nil)
	sessionCorrectionsRec := httptest.NewRecorder()

	engine.ServeHTTP(sessionCorrectionsRec, sessionCorrectionsReq)

	if sessionCorrectionsRec.Code != http.StatusOK {
		t.Fatalf("session corrections status code = %d, want %d; body = %s", sessionCorrectionsRec.Code, http.StatusOK, sessionCorrectionsRec.Body.String())
	}
	var sessionCorrectionsBody struct {
		Code int `json:"code"`
		Data []struct {
			MessageID int `json:"message_id"`
			SessionID int `json:"session_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(sessionCorrectionsRec.Body.Bytes(), &sessionCorrectionsBody); err != nil {
		t.Fatalf("session corrections response is not valid JSON: %v", err)
	}
	if sessionCorrectionsBody.Code != 0 {
		t.Fatalf("session corrections code = %d, want 0", sessionCorrectionsBody.Code)
	}
	if len(sessionCorrectionsBody.Data) != 1 {
		t.Fatalf("session corrections length = %d, want 1", len(sessionCorrectionsBody.Data))
	}
	if sessionCorrectionsBody.Data[0].MessageID != messageID {
		t.Fatalf("session correction message_id = %d, want %d", sessionCorrectionsBody.Data[0].MessageID, messageID)
	}

	scoreReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/scores", nil)
	scoreRec := httptest.NewRecorder()

	engine.ServeHTTP(scoreRec, scoreReq)

	if scoreRec.Code != http.StatusOK {
		t.Fatalf("score status code = %d, want %d; body = %s", scoreRec.Code, http.StatusOK, scoreRec.Body.String())
	}
	var scoreBody struct {
		Code int `json:"code"`
		Data struct {
			MessageID  int `json:"message_id"`
			SessionID  int `json:"session_id"`
			Grammar    int `json:"grammar"`
			Expression int `json:"expression"`
			TotalScore int `json:"total_score"`
		} `json:"data"`
	}
	if err := json.Unmarshal(scoreRec.Body.Bytes(), &scoreBody); err != nil {
		t.Fatalf("score response is not valid JSON: %v", err)
	}
	if scoreBody.Code != 0 {
		t.Fatalf("score code = %d, want 0", scoreBody.Code)
	}
	if scoreBody.Data.MessageID != messageID {
		t.Fatalf("score message_id = %d, want %d", scoreBody.Data.MessageID, messageID)
	}
	if scoreBody.Data.SessionID != sessionID {
		t.Fatalf("score session_id = %d, want %d", scoreBody.Data.SessionID, sessionID)
	}
	if scoreBody.Data.TotalScore != 77 {
		t.Fatalf("score total_score = %d, want 77", scoreBody.Data.TotalScore)
	}
	if scoreBody.Data.Grammar != 72 {
		t.Fatalf("score grammar = %d, want 72", scoreBody.Data.Grammar)
	}
	if scoreBody.Data.Expression != 80 {
		t.Fatalf("score expression = %d, want 80", scoreBody.Data.Expression)
	}
}

// TestMessageSendIncrementsTurnCountOnEachSuccessfulSend 验证多轮发送会持续递增轮次。
func TestMessageSendIncrementsTurnCountOnEachSuccessfulSend(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 2)

	first := postMessage(t, engine, sessionID, `{"content":"Could you recommend something light?"}`)
	if first.Data.TurnCount != 1 {
		t.Fatalf("first turn_count = %d, want 1", first.Data.TurnCount)
	}

	second := postMessage(t, engine, sessionID, `{"content":"I prefer chicken and no peanuts."}`)
	if second.Data.TurnCount != 2 {
		t.Fatalf("second turn_count = %d, want 2", second.Data.TurnCount)
	}
}

// TestMessageSendRejectsInvalidRequest 验证消息请求体非法或内容为空时返回统一错误。
func TestMessageSendRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantCode   int
		wantMsg    string
		wantStatus int
	}{
		{name: "invalid json", body: `{`, wantStatus: http.StatusBadRequest, wantCode: 3001, wantMsg: "invalid message request"},
		{name: "missing content", body: `{}`, wantStatus: http.StatusBadRequest, wantCode: 3001, wantMsg: "invalid message request"},
		{name: "blank content", body: `{"content":"   "}`, wantStatus: http.StatusBadRequest, wantCode: 3002, wantMsg: "message content is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			sessionID := createSession(t, engine, 1)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/sessions/"+strconv.Itoa(sessionID)+"/messages",
				bytes.NewBufferString(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assertErrorResponse(t, rec, tt.wantStatus, tt.wantCode, tt.wantMsg)
		})
	}
}

// TestMessageSendReturnsNotFound 验证不存在的 Session 不能发送消息。
func TestMessageSendReturnsNotFound(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/sessions/999/messages",
		bytes.NewBufferString(`{"content":"Hello"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, 2003, "session not found")
}

// TestMessageSendRejectsFinishedSession 验证已结束 Session 不允许继续发送消息。
func TestMessageSendRejectsFinishedSession(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	finishReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/finish", nil)
	finishRec := httptest.NewRecorder()
	engine.ServeHTTP(finishRec, finishReq)
	if finishRec.Code != http.StatusOK {
		t.Fatalf("finish status code = %d, want %d; body = %s", finishRec.Code, http.StatusOK, finishRec.Body.String())
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/sessions/"+strconv.Itoa(sessionID)+"/messages",
		bytes.NewBufferString(`{"content":"Can we continue?"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusConflict, 2004, "session already finished")
}

// TestSessionScoresRouteReturnsScoreNotFound 验证 Session 当前评分查询路由已注册并返回统一反馈错误。
func TestSessionScoresRouteReturnsScoreNotFound(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/scores", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, 4003, "score not found")
}

// TestMessageCorrectionsRouteReturnsCorrectionNotFound 验证单条消息纠错查询路由已注册并返回统一反馈错误。
func TestMessageCorrectionsRouteReturnsCorrectionNotFound(t *testing.T) {
	engine := New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/999/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, 4002, "correction not found")
}

// TestSessionCorrectionsRouteReturnsCorrectionNotFound 验证整场训练纠错查询路由已注册并返回统一反馈错误。
func TestSessionCorrectionsRouteReturnsCorrectionNotFound(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/corrections", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, 4002, "correction not found")
}

// TestReportRoutesGenerateAndReturnReport 验证结束训练后可以生成并重复查询课后报告。
func TestReportRoutesGenerateAndReturnReport(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)
	postMessage(t, engine, sessionID, `{"content":"I am study computer science and I have did a project."}`)

	finishReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/finish", nil)
	finishRec := httptest.NewRecorder()
	engine.ServeHTTP(finishRec, finishReq)
	if finishRec.Code != http.StatusOK {
		t.Fatalf("finish status code = %d, want %d; body = %s", finishRec.Code, http.StatusOK, finishRec.Body.String())
	}

	generateReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/report", nil)
	generateRec := httptest.NewRecorder()
	engine.ServeHTTP(generateRec, generateReq)

	if generateRec.Code != http.StatusOK {
		t.Fatalf("generate status code = %d, want %d; body = %s", generateRec.Code, http.StatusOK, generateRec.Body.String())
	}
	var generateBody struct {
		Code int `json:"code"`
		Data struct {
			SessionID        int      `json:"session_id"`
			TotalScore       int      `json:"total_score"`
			Summary          string   `json:"summary"`
			FrequentErrors   []string `json:"frequent_errors"`
			NextPracticePlan []string `json:"next_practice_plan"`
		} `json:"data"`
	}
	if err := json.Unmarshal(generateRec.Body.Bytes(), &generateBody); err != nil {
		t.Fatalf("generate response is not valid JSON: %v", err)
	}
	if generateBody.Code != 0 {
		t.Fatalf("generate code = %d, want 0", generateBody.Code)
	}
	if generateBody.Data.SessionID != sessionID {
		t.Fatalf("report session_id = %d, want %d", generateBody.Data.SessionID, sessionID)
	}
	if generateBody.Data.TotalScore != 77 {
		t.Fatalf("report total_score = %d, want 77", generateBody.Data.TotalScore)
	}
	if generateBody.Data.Summary == "" {
		t.Fatal("report summary is empty")
	}
	if len(generateBody.Data.FrequentErrors) == 0 {
		t.Fatal("report frequent_errors is empty")
	}
	if len(generateBody.Data.NextPracticePlan) == 0 {
		t.Fatal("report next_practice_plan is empty")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/report", nil)
	getRec := httptest.NewRecorder()
	engine.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get status code = %d, want %d; body = %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}
	var getBody struct {
		Code int `json:"code"`
		Data struct {
			SessionID int    `json:"session_id"`
			Summary   string `json:"summary"`
		} `json:"data"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("get response is not valid JSON: %v", err)
	}
	if getBody.Code != 0 {
		t.Fatalf("get code = %d, want 0", getBody.Code)
	}
	if getBody.Data.SessionID != sessionID {
		t.Fatalf("get session_id = %d, want %d", getBody.Data.SessionID, sessionID)
	}
	if getBody.Data.Summary != generateBody.Data.Summary {
		t.Fatalf("get summary = %q, want generated summary", getBody.Data.Summary)
	}
}

// TestReportRoutesReturnNotFoundBeforeGeneration 验证未生成报告时查询返回明确错误。
func TestReportRoutesReturnNotFoundBeforeGeneration(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/report", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusNotFound, 5003, "report not found")
}

// TestReportRoutesRequireFinishedSession 验证 running 状态不能生成课后报告。
func TestReportRoutesRequireFinishedSession(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)
	postMessage(t, engine, sessionID, `{"content":"I am study computer science."}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/report", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusConflict, 5002, "session not finished")
}

// TestReportRoutesRequireFeedbackData 验证缺少纠错或评分时报告生成失败。
func TestReportRoutesRequireFeedbackData(t *testing.T) {
	engine := New()
	sessionID := createSession(t, engine, 1)

	finishReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/finish", nil)
	finishRec := httptest.NewRecorder()
	engine.ServeHTTP(finishRec, finishReq)
	if finishRec.Code != http.StatusOK {
		t.Fatalf("finish status code = %d, want %d; body = %s", finishRec.Code, http.StatusOK, finishRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/report", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertErrorResponse(t, rec, http.StatusConflict, 5004, "report feedback missing")
}

// TestSessionHistoryRoutesReturnPaginatedSummaries 验证历史记录列表支持分页、用户过滤和报告状态摘要。
func TestSessionHistoryRoutesReturnPaginatedSummaries(t *testing.T) {
	engine := New()
	reportedSessionID := createSessionForUser(t, engine, 1, 42)
	postMessage(t, engine, reportedSessionID, `{"content":"I am study computer science and I have did a project."}`)
	finishSession(t, engine, reportedSessionID)
	generateReport(t, engine, reportedSessionID)

	otherSessionID := createSessionForUser(t, engine, 2, 7)
	postMessage(t, engine, otherSessionID, `{"content":"Could you recommend something light?"}`)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?page=1&page_size=1", nil)
	listRec := httptest.NewRecorder()
	engine.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("history list status code = %d, want %d; body = %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	var listBody sessionHistoryListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("history list response is not valid JSON: %v", err)
	}
	if listBody.Code != 0 {
		t.Fatalf("history list code = %d, want 0", listBody.Code)
	}
	if listBody.Data.Page != 1 {
		t.Fatalf("history page = %d, want 1", listBody.Data.Page)
	}
	if listBody.Data.PageSize != 1 {
		t.Fatalf("history page_size = %d, want 1", listBody.Data.PageSize)
	}
	if listBody.Data.Total != 2 {
		t.Fatalf("history total = %d, want 2", listBody.Data.Total)
	}
	if len(listBody.Data.Items) != 1 {
		t.Fatalf("history items length = %d, want 1", len(listBody.Data.Items))
	}

	userReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/42/sessions?page=1&page_size=10", nil)
	userRec := httptest.NewRecorder()
	engine.ServeHTTP(userRec, userReq)

	if userRec.Code != http.StatusOK {
		t.Fatalf("user history status code = %d, want %d; body = %s", userRec.Code, http.StatusOK, userRec.Body.String())
	}
	var userBody sessionHistoryListResponse
	if err := json.Unmarshal(userRec.Body.Bytes(), &userBody); err != nil {
		t.Fatalf("user history response is not valid JSON: %v", err)
	}
	if userBody.Data.Total != 1 {
		t.Fatalf("user history total = %d, want 1", userBody.Data.Total)
	}
	if len(userBody.Data.Items) != 1 {
		t.Fatalf("user history items length = %d, want 1", len(userBody.Data.Items))
	}
	item := userBody.Data.Items[0]
	if item.SessionID != reportedSessionID {
		t.Fatalf("history session_id = %d, want %d", item.SessionID, reportedSessionID)
	}
	if item.UserID != 42 {
		t.Fatalf("history user_id = %d, want 42", item.UserID)
	}
	if item.Scenario.Code != "interview" {
		t.Fatalf("history scenario code = %q, want interview", item.Scenario.Code)
	}
	if item.TurnCount != 1 {
		t.Fatalf("history turn_count = %d, want 1", item.TurnCount)
	}
	if item.TotalScore == nil || *item.TotalScore != 77 {
		t.Fatalf("history total_score = %v, want 77", item.TotalScore)
	}
	if item.ReportStatus != "generated" {
		t.Fatalf("history report_status = %q, want generated", item.ReportStatus)
	}
	assertRFC3339(t, item.CreatedAt)
	if item.EndedAt == nil {
		t.Fatal("history ended_at = nil, want finished timestamp")
	}
	assertRFC3339(t, *item.EndedAt)
}

// TestSessionHistoryRoutesValidatePaginationAndUserID 验证历史记录路由对分页和用户 ID 做基础校验。
func TestSessionHistoryRoutesValidatePaginationAndUserID(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "invalid page", path: "/api/v1/sessions?page=0&page_size=10"},
		{name: "invalid page size", path: "/api/v1/sessions?page=1&page_size=0"},
		{name: "invalid user id", path: "/api/v1/users/abc/sessions?page=1&page_size=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assertErrorResponse(t, rec, http.StatusBadRequest, 6001, "invalid history request")
		})
	}
}

type sessionHistoryListResponse struct {
	Code int `json:"code"`
	Data struct {
		Items    []sessionHistoryItem `json:"items"`
		Page     int                  `json:"page"`
		PageSize int                  `json:"page_size"`
		Total    int                  `json:"total"`
	} `json:"data"`
}

type sessionHistoryItem struct {
	SessionID    int             `json:"session_id"`
	SessionNo    string          `json:"session_no"`
	UserID       int             `json:"user_id"`
	Scenario     historyScenario `json:"scenario"`
	Status       string          `json:"status"`
	TurnCount    int             `json:"turn_count"`
	TotalScore   *int            `json:"total_score"`
	ReportStatus string          `json:"report_status"`
	CreatedAt    string          `json:"created_at"`
	EndedAt      *string         `json:"ended_at"`
}

type historyScenario struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Difficulty  string `json:"difficulty"`
}

type messagePayload struct {
	ID        int    `json:"id"`
	SessionID int    `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Stage     string `json:"stage"`
	CreatedAt string `json:"created_at"`
}

type postMessageResponse struct {
	Data struct {
		UserMessage       messagePayload    `json:"user_message"`
		AIMessage         messagePayload    `json:"ai_message"`
		Stage             string            `json:"stage"`
		NextGoal          string            `json:"next_goal"`
		TurnCount         int               `json:"turn_count"`
		CorrectionSummary correctionSummary `json:"correction_summary"`
		ScoreSummary      scoreSummary      `json:"score_summary"`
	} `json:"data"`
}

type correctionSummary struct {
	HasErrors  bool `json:"has_errors"`
	ErrorCount int  `json:"error_count"`
}

type scoreSummary struct {
	TotalScore int `json:"total_score"`
	Grammar    int `json:"grammar"`
	Expression int `json:"expression"`
}

func postMessage(t *testing.T, engine http.Handler, sessionID int, body string) postMessageResponse {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/sessions/"+strconv.Itoa(sessionID)+"/messages",
		bytes.NewBufferString(body),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var parsed postMessageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	return parsed
}

func assertRFC3339(t *testing.T, value string) {
	t.Helper()

	if _, err := time.Parse(time.RFC3339, value); err != nil {
		t.Fatalf("time %q is not RFC3339: %v", value, err)
	}
}

func createSession(t *testing.T, engine http.Handler, scenarioID int) int {
	t.Helper()

	return createSessionForUser(t, engine, scenarioID, 1)
}

func createSessionForUser(t *testing.T, engine http.Handler, scenarioID int, userID int) int {
	t.Helper()

	requestBody := `{"scenario_id":` + strconv.Itoa(scenarioID) + `,"user_id":` + strconv.Itoa(userID) + `}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("create session status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		Data struct {
			SessionID int `json:"session_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("create session response is not valid JSON: %v", err)
	}
	if body.Data.SessionID <= 0 {
		t.Fatalf("session_id = %d, want positive", body.Data.SessionID)
	}

	return body.Data.SessionID
}

func finishSession(t *testing.T, engine http.Handler, sessionID int) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/finish", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("finish session status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func generateReport(t *testing.T, engine http.Handler, sessionID int) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+strconv.Itoa(sessionID)+"/report", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("generate report status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, httpStatus int, code int, message string) {
	t.Helper()

	if rec.Code != httpStatus {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, httpStatus, rec.Body.String())
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if body.Code != code {
		t.Fatalf("code = %d, want %d", body.Code, code)
	}
	if body.Message != message {
		t.Fatalf("message = %q, want %q", body.Message, message)
	}
}
