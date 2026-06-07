package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"speakmate/internal/config"
	"speakmate/internal/response"
	"speakmate/internal/service"
)

const (
	audioWSEventStart             = "start"
	audioWSEventAudioChunk        = "audio_chunk"
	audioWSEventPartialTranscript = "partial_transcript"
	audioWSEventFinalTranscript   = "final_transcript"
	audioWSEventCorrection        = "correction"
	audioWSEventScoreUpdated      = "score_updated"
	audioWSEventEnd               = "end"
	audioWSEventError             = "error"
	audioWSUnavailableCode        = 7101
	audioWSWriteTimeout           = 5 * time.Second
)

// AudioWebSocketHandler 负责处理实时音频 WebSocket API。
type AudioWebSocketHandler struct {
	service  *service.AudioStreamService
	upgrader websocket.Upgrader
}

// NewAudioWebSocketHandler 创建实时音频 WebSocket Handler。
func NewAudioWebSocketHandler(service *service.AudioStreamService, corsConfigs ...config.CORSConfig) *AudioWebSocketHandler {
	checkOrigin := func(r *http.Request) bool {
		return true
	}
	if len(corsConfigs) > 0 {
		checkOrigin = webSocketOriginChecker(corsConfigs[0])
	}

	return &AudioWebSocketHandler{
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: checkOrigin,
		},
	}
}

// Stream 接收浏览器音频分片，推送 partial/final transcript 和反馈事件。
func (h *AudioWebSocketHandler) Stream(c *gin.Context) {
	id, ok := parsePositiveSessionID(c)
	if !ok {
		return
	}
	if h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, audioWSUnavailableCode, "audio websocket unavailable")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	closedByHandler := false
	defer func() {
		if !closedByHandler {
			_ = h.service.RecordConnectionClosed(c.Request.Context(), id, "client_disconnect")
		}
		_ = conn.Close()
	}()

	var streamSession *service.AudioStream
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			_ = h.service.RecordConnectionClosed(c.Request.Context(), id, "read_closed")
			return
		}

		switch messageType {
		case websocket.TextMessage:
			nextStream, shouldClose := h.handleTextMessage(c, conn, id, streamSession, message)
			if nextStream != nil {
				streamSession = nextStream
			}
			if shouldClose {
				closedByHandler = true
				_ = h.service.RecordConnectionClosed(c.Request.Context(), id, "handler_close")
				return
			}
		case websocket.BinaryMessage:
			if streamSession == nil {
				if !h.writeError(c, conn, id, service.ErrInvalidAudioRequest) {
					return
				}
				continue
			}
			if !h.writePartial(conn, id, streamSession, service.AudioStreamChunkInput{
				Audio:   message,
				Context: c.Request.Context(),
			}) {
				return
			}
		default:
			if !h.writeError(c, conn, id, service.ErrInvalidAudioRequest) {
				return
			}
		}
	}
}

func (h *AudioWebSocketHandler) handleTextMessage(c *gin.Context, conn *websocket.Conn, sessionID int, streamSession *service.AudioStream, message []byte) (*service.AudioStream, bool) {
	var event audioWSClientEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return nil, !h.writeError(c, conn, sessionID, service.ErrInvalidAudioRequest)
	}

	switch event.Type {
	case audioWSEventStart:
		started, err := h.service.Start(service.StartAudioStreamInput{
			SessionID:   sessionID,
			ContentType: event.Payload.ContentType,
		})
		if err != nil {
			return nil, !h.writeError(c, conn, sessionID, err)
		}
		if !writeAudioWSEvent(conn, sessionID, audioWSEventStart, audioWSStartPayload{
			ContentType: event.Payload.ContentType,
		}) {
			return nil, true
		}

		return started, false
	case audioWSEventAudioChunk:
		if streamSession == nil {
			return nil, !h.writeError(c, conn, sessionID, service.ErrInvalidAudioRequest)
		}
		audio, err := decodeAudioChunk(event.Payload.AudioBase64)
		if err != nil {
			return nil, !h.writeError(c, conn, sessionID, err)
		}

		return nil, !h.writePartial(conn, sessionID, streamSession, service.AudioStreamChunkInput{
			Audio:    audio,
			Sequence: event.Payload.Sequence,
			Context:  c.Request.Context(),
		})
	case audioWSEventEnd:
		if streamSession == nil {
			return nil, !h.writeError(c, conn, sessionID, service.ErrInvalidAudioRequest)
		}
		result, err := streamSession.Finish(c.Request.Context())
		if err != nil {
			return nil, !h.writeError(c, conn, sessionID, err)
		}
		if !writeAudioWSEvent(conn, sessionID, audioWSEventFinalTranscript, finalTranscriptPayload(result)) {
			return nil, true
		}
		if !writeAudioWSEvent(conn, sessionID, audioWSEventCorrection, toCorrectionSummaryResponse(result.CorrectionSummary)) {
			return nil, true
		}
		if !writeAudioWSEvent(conn, sessionID, audioWSEventScoreUpdated, toScoreSummaryResponse(result.ScoreSummary)) {
			return nil, true
		}
		if !writeAudioWSEvent(conn, sessionID, audioWSEventEnd, audioWSEndPayload{Reason: "client_end"}) {
			return nil, true
		}
		writeAudioWSClose(conn, websocket.CloseNormalClosure, "end")

		return nil, true
	default:
		return nil, !h.writeError(c, conn, sessionID, service.ErrInvalidAudioRequest)
	}
}

func (h *AudioWebSocketHandler) writePartial(conn *websocket.Conn, sessionID int, streamSession *service.AudioStream, input service.AudioStreamChunkInput) bool {
	partial, err := streamSession.AppendChunk(input)
	if err != nil {
		return h.writeError(nil, conn, sessionID, err)
	}

	return writeAudioWSEvent(conn, sessionID, audioWSEventPartialTranscript, partialTranscriptPayload{
		Transcript: partial.Transcript,
		Sequence:   partial.Sequence,
	})
}

func (h *AudioWebSocketHandler) writeError(c *gin.Context, conn *websocket.Conn, sessionID int, err error) bool {
	if h.service != nil {
		ctx := contextFromGin(c)
		if stateErr := h.service.RecordConnectionError(ctx, sessionID, err); stateErr != nil {
			return writeAudioWSError(conn, sessionID, stateErr)
		}
	}

	return writeAudioWSError(conn, sessionID, err)
}

type audioWSClientEvent struct {
	Type    string               `json:"type"`
	Payload audioWSClientPayload `json:"payload"`
}

type audioWSClientPayload struct {
	ContentType string `json:"content_type"`
	AudioBase64 string `json:"audio_base64"`
	Sequence    int    `json:"sequence"`
}

type audioWSServerEvent struct {
	Type      string    `json:"type"`
	SessionID int       `json:"session_id"`
	Payload   any       `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type audioWSStartPayload struct {
	ContentType string `json:"content_type"`
}

type partialTranscriptPayload struct {
	Transcript string `json:"transcript"`
	Sequence   int    `json:"sequence"`
}

type audioWSEndPayload struct {
	Reason string `json:"reason"`
}

type audioWSErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type audioWSFinalTranscriptPayload struct {
	Transcript  string          `json:"transcript"`
	UserMessage messageResponse `json:"user_message"`
	AIMessage   messageResponse `json:"ai_message"`
	Stage       string          `json:"stage"`
	NextGoal    string          `json:"next_goal"`
	TurnCount   int             `json:"turn_count"`
}

func finalTranscriptPayload(result service.AudioStreamResult) audioWSFinalTranscriptPayload {
	return audioWSFinalTranscriptPayload{
		Transcript:  result.Transcript,
		UserMessage: toMessageResponse(result.UserMessage),
		AIMessage:   toMessageResponse(result.AIMessage),
		Stage:       result.Stage,
		NextGoal:    result.NextGoal,
		TurnCount:   result.TurnCount,
	}
}

func decodeAudioChunk(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, service.ErrAudioFileRequired
	}

	audio, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, service.ErrInvalidAudioRequest
	}

	return audio, nil
}

func writeAudioWSEvent(conn *websocket.Conn, sessionID int, eventType string, payload any) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(audioWSWriteTimeout)); err != nil {
		return false
	}

	return conn.WriteJSON(audioWSServerEvent{
		Type:      eventType,
		SessionID: sessionID,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}) == nil
}

func writeAudioWSClose(conn *websocket.Conn, code int, text string) {
	deadline := time.Now().Add(audioWSWriteTimeout)
	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, text), deadline)
}

func writeAudioWSError(conn *websocket.Conn, sessionID int, err error) bool {
	code, message := audioWSError(err)

	return writeAudioWSEvent(conn, sessionID, audioWSEventError, audioWSErrorPayload{
		Code:    code,
		Message: message,
	})
}

func audioWSError(err error) (string, string) {
	switch {
	case errors.Is(err, service.ErrInvalidAudioRequest):
		return "invalid_audio_request", "invalid audio request"
	case errors.Is(err, service.ErrAudioFileRequired):
		return "audio_file_required", "audio file is required"
	case errors.Is(err, service.ErrAudioFileTooLarge):
		return "audio_file_too_large", "audio file too large"
	case errors.Is(err, service.ErrAudioFileTypeUnsupported):
		return "audio_file_type_unsupported", "audio file type unsupported"
	case errors.Is(err, service.ErrASRClientFailed):
		return "asr_client_failed", "asr client failed"
	case errors.Is(err, service.ErrAudioTranscriptRequired):
		return "audio_transcript_required", "audio transcript is required"
	case errors.Is(err, service.ErrStateStoreFailed):
		return "session_state_store_failed", "session state store failed"
	case errors.Is(err, service.ErrSessionNotFound):
		return "session_not_found", "session not found"
	case errors.Is(err, service.ErrSessionAlreadyFinished):
		return "session_already_finished", "session already finished"
	case errors.Is(err, service.ErrMessageContentRequired):
		return "message_content_required", "message content is required"
	default:
		return "audio_websocket_failed", "audio websocket failed"
	}
}

func contextFromGin(c *gin.Context) context.Context {
	if c == nil || c.Request == nil {
		return context.Background()
	}

	return c.Request.Context()
}

func webSocketOriginChecker(cfg config.CORSConfig) func(*http.Request) bool {
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	allowWildcard := false
	for _, origin := range cfg.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowWildcard = true
			continue
		}
		allowedOrigins[origin] = struct{}{}
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		if allowWildcard {
			return true
		}
		_, ok := allowedOrigins[origin]
		return ok
	}
}
