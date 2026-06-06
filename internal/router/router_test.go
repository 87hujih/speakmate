package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
