package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/config"
)

func TestCORSMiddlewareAllowsConfiguredOriginAndPreflight(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(CORS(config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}))
	engine.GET("/api/v1/scenarios", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/scenarios", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want configured origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "OPTIONS") {
		t.Fatalf("Access-Control-Allow-Methods = %q, want OPTIONS", got)
	}
}

func TestRequestLoggerRedactsSensitiveQueryValues(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	var output bytes.Buffer
	engine := gin.New()
	engine.Use(RequestLogger(log.New(&output, "", 0)))
	engine.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/health?api_key=test-key&password=test-password&q=visible", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	logLine := output.String()
	for _, secret := range []string{"test-key", "test-password"} {
		if strings.Contains(logLine, secret) {
			t.Fatalf("request log %q still contains secret %q", logLine, secret)
		}
	}
	if !strings.Contains(logLine, "q=visible") {
		t.Fatalf("request log = %q, want non-sensitive query value preserved", logLine)
	}
}

func TestRecoverMiddlewareReturnsUnifiedErrorAndRedactsPanic(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	var output bytes.Buffer
	engine := gin.New()
	engine.Use(Recover(log.New(&output, "", 0)))
	engine.GET("/panic", func(c *gin.Context) {
		panic("password=panic-secret")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic?token=query-secret", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Code != 9001 || body.Message != "internal server error" {
		t.Fatalf("body = %#v, want code 9001 internal server error", body)
	}

	logLine := output.String()
	for _, secret := range []string{"panic-secret", "query-secret"} {
		if strings.Contains(logLine, secret) {
			t.Fatalf("recover log %q still contains secret %q", logLine, secret)
		}
	}
}

func TestRequestTimeoutWritesGatewayTimeoutWhenHandlerHonorsContext(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(RequestTimeout(10 * time.Millisecond))
	engine.GET("/slow", func(c *gin.Context) {
		<-c.Request.Context().Done()
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusGatewayTimeout, rec.Body.String())
	}
}

func TestRequestTimeoutSkipsLongLivedStreamRoutes(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(RequestTimeout(10 * time.Millisecond))
	engine.GET("/api/v1/sessions/:id/stream", func(c *gin.Context) {
		select {
		case <-time.After(20 * time.Millisecond):
			c.String(http.StatusOK, "stream alive")
		case <-c.Request.Context().Done():
			c.String(599, "context cancelled")
		}
	})
	engine.GET("/api/v1/sessions/:id/audio/ws", func(c *gin.Context) {
		select {
		case <-time.After(20 * time.Millisecond):
			c.String(http.StatusOK, "ws alive")
		case <-c.Request.Context().Done():
			c.String(599, "context cancelled")
		}
	})

	for _, path := range []string{"/api/v1/sessions/1/stream", "/api/v1/sessions/1/audio/ws"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s status code = %d, want %d; body = %s", path, rec.Code, http.StatusOK, rec.Body.String())
		}
	}
}
