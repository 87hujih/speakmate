package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/model"
	"speakmate/internal/service"
)

func TestAudioHandlerUploadsMultipartAudio(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	audioService := &fakeAudioService{
		result: sampleAudioUploadResult("I built a robot control project."),
	}
	audioHandler := NewAudioHandler(audioService)
	engine := gin.New()
	engine.POST("/api/v1/sessions/:id/audio", audioHandler.Upload)

	body, contentType := multipartAudioBody(t, "answer.webm", "audio/webm", []byte{0x01, 0x02})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/7/audio", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if audioService.input.SessionID != 7 {
		t.Fatalf("session id = %d, want 7", audioService.input.SessionID)
	}
	if audioService.input.Filename != "answer.webm" {
		t.Fatalf("filename = %q, want answer.webm", audioService.input.Filename)
	}
	if audioService.input.ContentType != "audio/webm" {
		t.Fatalf("content type = %q, want audio/webm", audioService.input.ContentType)
	}
	if !bytes.Equal(audioService.input.Audio, []byte{0x01, 0x02}) {
		t.Fatalf("audio bytes = %v, want uploaded bytes", audioService.input.Audio)
	}

	var parsed struct {
		Code int `json:"code"`
		Data struct {
			Transcript  string          `json:"transcript"`
			UserMessage messageResponse `json:"user_message"`
			AIMessage   messageResponse `json:"ai_message"`
			TurnCount   int             `json:"turn_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if parsed.Code != 0 {
		t.Fatalf("code = %d, want 0", parsed.Code)
	}
	if parsed.Data.Transcript != "I built a robot control project." {
		t.Fatalf("transcript = %q, want uploaded transcript", parsed.Data.Transcript)
	}
	if parsed.Data.UserMessage.Role != "user" {
		t.Fatalf("user message role = %q, want user", parsed.Data.UserMessage.Role)
	}
	if parsed.Data.AIMessage.Role != "ai" {
		t.Fatalf("ai message role = %q, want ai", parsed.Data.AIMessage.Role)
	}
	if parsed.Data.TurnCount != 1 {
		t.Fatalf("turn count = %d, want 1", parsed.Data.TurnCount)
	}
}

func TestAudioHandlerRequiresMultipartAudioFile(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	audioService := &fakeAudioService{}
	audioHandler := NewAudioHandler(audioService)
	engine := gin.New()
	engine.POST("/api/v1/sessions/:id/audio", audioHandler.Upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close returned error: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/7/audio", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	assertAudioErrorResponse(t, rec, http.StatusBadRequest, 7002, "audio file is required")
	if audioService.callCount != 0 {
		t.Fatalf("call count = %d, want 0", audioService.callCount)
	}
}

func TestAudioHandlerMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
		wantMsg    string
	}{
		{name: "unsupported type", err: service.ErrAudioFileTypeUnsupported, wantStatus: http.StatusBadRequest, wantCode: 7004, wantMsg: "audio file type unsupported"},
		{name: "too large", err: service.ErrAudioFileTooLarge, wantStatus: http.StatusRequestEntityTooLarge, wantCode: 7003, wantMsg: "audio file too large"},
		{name: "blank transcript", err: service.ErrAudioTranscriptRequired, wantStatus: http.StatusBadRequest, wantCode: 7006, wantMsg: "audio transcript is required"},
		{name: "asr failed", err: service.ErrASRClientFailed, wantStatus: http.StatusBadGateway, wantCode: 7005, wantMsg: "asr client failed"},
		{name: "session finished", err: service.ErrSessionAlreadyFinished, wantStatus: http.StatusConflict, wantCode: 2004, wantMsg: "session already finished"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.ReleaseMode)
			audioService := &fakeAudioService{err: tt.err}
			audioHandler := NewAudioHandler(audioService)
			engine := gin.New()
			engine.POST("/api/v1/sessions/:id/audio", audioHandler.Upload)

			body, contentType := multipartAudioBody(t, "answer.webm", "audio/webm", []byte{0x01})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/7/audio", body)
			req.Header.Set("Content-Type", contentType)
			rec := httptest.NewRecorder()

			engine.ServeHTTP(rec, req)

			assertAudioErrorResponse(t, rec, tt.wantStatus, tt.wantCode, tt.wantMsg)
		})
	}
}

type fakeAudioService struct {
	result    service.UploadAudioResult
	err       error
	callCount int
	input     service.UploadAudioInput
}

func (s *fakeAudioService) UploadAudio(input service.UploadAudioInput) (service.UploadAudioResult, error) {
	s.callCount++
	s.input = input
	if s.err != nil {
		return service.UploadAudioResult{}, s.err
	}

	return s.result, nil
}

func sampleAudioUploadResult(transcript string) service.UploadAudioResult {
	createdAt := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)

	return service.UploadAudioResult{
		Transcript: transcript,
		SendMessageResult: service.SendMessageResult{
			UserMessage: model.Message{
				ID:        11,
				SessionID: 7,
				Role:      model.MessageRoleUser,
				Content:   transcript,
				Stage:     "项目经历",
				CreatedAt: createdAt,
			},
			AIMessage: model.Message{
				ID:        12,
				SessionID: 7,
				Role:      model.MessageRoleAI,
				Content:   "Could you explain one technical challenge?",
				Stage:     "技术追问",
				CreatedAt: createdAt,
			},
			Stage:     "技术追问",
			NextGoal:  "ask user to explain a technical challenge",
			TurnCount: 1,
			CorrectionSummary: service.CorrectionSummary{
				HasErrors:  false,
				ErrorCount: 0,
			},
			ScoreSummary: service.ScoreSummary{
				TotalScore: 86,
				Grammar:    88,
				Expression: 84,
			},
		},
	}
}

func multipartAudioBody(t *testing.T, filename string, contentType string, data []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="audio"; filename="` + filename + `"`},
		"Content-Type":        {contentType},
	})
	if err != nil {
		t.Fatalf("CreatePart returned error: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("part.Write returned error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close returned error: %v", err)
	}

	return body, writer.FormDataContentType()
}

func assertAudioErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, httpStatus int, code int, message string) {
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
