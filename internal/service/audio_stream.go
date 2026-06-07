package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"speakmate/internal/agent"
	"speakmate/internal/state"
)

const defaultAudioStreamContentType = "audio/webm"

// AudioStreamService 封装 WebSocket 音频分片、Mock 转写和消息链路复用。
type AudioStreamService struct {
	messageSender      AudioMessageSender
	asr                agent.ASRClient
	stateStore         state.SessionStateStore
	maxBytes           int
	transcribePartials bool
}

type AudioStreamOption func(*AudioStreamService)

// WithMaxAudioStreamBytes 覆盖 WebSocket 音频流大小上限。
func WithMaxAudioStreamBytes(maxBytes int) AudioStreamOption {
	return func(service *AudioStreamService) {
		if maxBytes > 0 {
			service.maxBytes = maxBytes
		}
	}
}

// WithAudioStreamPartialTranscription 控制 AppendChunk 是否调用 ASR 生成 partial。
func WithAudioStreamPartialTranscription(enabled bool) AudioStreamOption {
	return func(service *AudioStreamService) {
		service.transcribePartials = enabled
	}
}

// WithAudioStreamStateStore 设置 WebSocket 连接状态存储。
func WithAudioStreamStateStore(store state.SessionStateStore) AudioStreamOption {
	return func(service *AudioStreamService) {
		if store != nil {
			service.stateStore = store
		}
	}
}

// NewAudioStreamService 创建实时音频流服务。
func NewAudioStreamService(messageSender AudioMessageSender, asr agent.ASRClient, opts ...AudioStreamOption) *AudioStreamService {
	service := &AudioStreamService{
		messageSender:      messageSender,
		asr:                asr,
		maxBytes:           MaxAudioUploadBytes,
		transcribePartials: true,
	}
	if service.asr == nil {
		service.asr = agent.NewMockASRClient()
	}
	for _, opt := range opts {
		opt(service)
	}

	return service
}

// StartAudioStreamInput 是一次 WebSocket 音频流的起始参数。
type StartAudioStreamInput struct {
	SessionID   int
	ContentType string
}

// AudioStream 表示单条 WebSocket 连接内的音频分片状态。
type AudioStream struct {
	service     *AudioStreamService
	sessionID   int
	contentType string
	audio       []byte
	chunkCount  int
	ended       bool
}

// AudioStreamChunkInput 是一段实时音频分片。
type AudioStreamChunkInput struct {
	Audio    []byte
	Sequence int
	Context  context.Context
}

// AudioStreamPartialResult 是一次 partial transcript 输出。
type AudioStreamPartialResult struct {
	Transcript string
	Sequence   int
}

// AudioStreamResult 是 final transcript 进入训练链路后的输出。
type AudioStreamResult struct {
	Transcript string
	SendMessageResult
}

// Start 创建连接级音频流状态。
func (s *AudioStreamService) Start(input StartAudioStreamInput) (*AudioStream, error) {
	if input.SessionID <= 0 {
		return nil, ErrInvalidAudioRequest
	}

	contentType := normalizeAudioStreamContentType(input.ContentType)
	if !isSupportedAudioContentType(contentType) {
		return nil, ErrAudioFileTypeUnsupported
	}

	stream := &AudioStream{
		service:     s,
		sessionID:   input.SessionID,
		contentType: contentType,
		audio:       []byte{},
	}
	if err := s.saveConnectionState(context.Background(), state.WebSocketConnectionState{
		SessionID:   input.SessionID,
		Status:      "started",
		ContentType: contentType,
		UpdatedAt:   timeNowUTC(),
	}); err != nil {
		return nil, err
	}

	return stream, nil
}

// AppendChunk 保存一个音频分片并返回稳定的 Mock partial transcript。
func (s *AudioStream) AppendChunk(input AudioStreamChunkInput) (AudioStreamPartialResult, error) {
	if s == nil || s.service == nil || s.ended {
		return AudioStreamPartialResult{}, ErrInvalidAudioRequest
	}
	if len(input.Audio) == 0 {
		return AudioStreamPartialResult{}, ErrAudioFileRequired
	}
	if len(s.audio)+len(input.Audio) > s.service.maxBytes {
		return AudioStreamPartialResult{}, ErrAudioFileTooLarge
	}

	s.audio = append(s.audio, input.Audio...)
	s.chunkCount++
	sequence := input.Sequence
	if sequence <= 0 {
		sequence = s.chunkCount
	}
	if err := s.service.saveConnectionState(input.Context, state.WebSocketConnectionState{
		SessionID:    s.sessionID,
		Status:       "receiving",
		ContentType:  s.contentType,
		ChunkCount:   s.chunkCount,
		LastSequence: sequence,
		UpdatedAt:    timeNowUTC(),
	}); err != nil {
		return AudioStreamPartialResult{}, err
	}

	if !s.service.transcribePartials {
		return AudioStreamPartialResult{
			Transcript: "",
			Sequence:   sequence,
		}, nil
	}

	output, err := s.service.transcribe(input.Context, s.audio, "stream-partial", s.contentType)
	if err != nil {
		return AudioStreamPartialResult{}, err
	}

	return AudioStreamPartialResult{
		Transcript: partialTranscript(output.Transcript, sequence),
		Sequence:   sequence,
	}, nil
}

// Finish 完成音频流，生成 final transcript，并复用 SendMessage。
func (s *AudioStream) Finish(ctx context.Context) (AudioStreamResult, error) {
	if s == nil || s.service == nil || s.ended {
		return AudioStreamResult{}, ErrInvalidAudioRequest
	}
	s.ended = true
	if len(s.audio) == 0 {
		return AudioStreamResult{}, ErrAudioFileRequired
	}

	output, err := s.service.transcribe(ctx, s.audio, "stream-final", s.contentType)
	if err != nil {
		return AudioStreamResult{}, err
	}

	transcript := strings.TrimSpace(output.Transcript)
	if transcript == "" {
		return AudioStreamResult{}, ErrAudioTranscriptRequired
	}
	if s.service.messageSender == nil {
		return AudioStreamResult{}, ErrInvalidAudioRequest
	}

	result, err := s.service.messageSender.SendMessage(SendMessageInput{
		SessionID: s.sessionID,
		Content:   transcript,
		Context:   ctx,
	})
	if err != nil {
		return AudioStreamResult{}, err
	}
	if err := s.service.saveConnectionState(ctx, state.WebSocketConnectionState{
		SessionID:    s.sessionID,
		Status:       "ended",
		ContentType:  s.contentType,
		ChunkCount:   s.chunkCount,
		LastSequence: s.chunkCount,
		UpdatedAt:    timeNowUTC(),
	}); err != nil {
		return AudioStreamResult{}, err
	}

	return AudioStreamResult{
		Transcript:        transcript,
		SendMessageResult: result,
	}, nil
}

// RecordConnectionError 写入 WebSocket 连接错误状态。
func (s *AudioStreamService) RecordConnectionError(ctx context.Context, sessionID int, err error) error {
	lastError := ""
	if err != nil {
		lastError = err.Error()
	}

	return s.saveConnectionState(ctx, state.WebSocketConnectionState{
		SessionID: sessionID,
		Status:    "error",
		LastError: lastError,
		UpdatedAt: timeNowUTC(),
	})
}

// RecordConnectionClosed 写入 WebSocket 连接关闭状态。
func (s *AudioStreamService) RecordConnectionClosed(ctx context.Context, sessionID int, reason string) error {
	return s.saveConnectionState(ctx, state.WebSocketConnectionState{
		SessionID: sessionID,
		Status:    "closed",
		LastError: reason,
		UpdatedAt: timeNowUTC(),
	})
}

func (s *AudioStreamService) saveConnectionState(ctx context.Context, connection state.WebSocketConnectionState) error {
	if s.stateStore == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := s.stateStore.SaveWebSocketConnection(ctx, connection); err != nil {
		return fmt.Errorf("%w: save websocket connection: %v", ErrStateStoreFailed, err)
	}

	return nil
}

func (s *AudioStreamService) transcribe(ctx context.Context, audio []byte, filename string, contentType string) (agent.ASROutput, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	asrClient := s.asr
	if asrClient == nil {
		asrClient = agent.NewMockASRClient()
	}

	output, err := asrClient.Transcribe(ctx, agent.ASRInput{
		Filename:    fmt.Sprintf("%s.webm", filename),
		ContentType: contentType,
		Audio:       audio,
	})
	if err != nil {
		return agent.ASROutput{}, fmt.Errorf("%w: %v", ErrASRClientFailed, err)
	}

	return output, nil
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}

func normalizeAudioStreamContentType(contentType string) string {
	trimmed := strings.TrimSpace(contentType)
	if trimmed == "" {
		return defaultAudioStreamContentType
	}

	return trimmed
}

func partialTranscript(transcript string, sequence int) string {
	words := strings.Fields(transcript)
	if len(words) == 0 {
		return ""
	}
	if len(words) <= 3 {
		return strings.Join(words, " ")
	}

	wordCount := sequence * 3
	if wordCount >= len(words) {
		wordCount = len(words) - 1
	}
	if wordCount < 1 {
		wordCount = 1
	}

	return strings.Join(words[:wordCount], " ")
}
