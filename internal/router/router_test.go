package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

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

func createSession(t *testing.T, engine http.Handler, scenarioID int) int {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(`{"scenario_id":`+strconv.Itoa(scenarioID)+`}`))
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
