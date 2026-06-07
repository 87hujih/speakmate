package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"speakmate/internal/agent"
)

const (
	// MaxAudioUploadBytes 是单段音频上传的默认大小上限。
	MaxAudioUploadBytes = 10 * 1024 * 1024
)

var (
	// ErrInvalidAudioRequest 表示音频上传业务参数非法。
	ErrInvalidAudioRequest = errors.New("invalid audio request")
	// ErrAudioFileRequired 表示请求中缺少音频文件或文件为空。
	ErrAudioFileRequired = errors.New("audio file is required")
	// ErrAudioFileTooLarge 表示音频文件超过大小上限。
	ErrAudioFileTooLarge = errors.New("audio file too large")
	// ErrAudioFileTypeUnsupported 表示音频类型不在当前支持范围内。
	ErrAudioFileTypeUnsupported = errors.New("audio file type unsupported")
	// ErrASRClientFailed 表示 ASR Client 转写失败。
	ErrASRClientFailed = errors.New("asr client failed")
	// ErrAudioTranscriptRequired 表示 ASR 没有返回有效文本。
	ErrAudioTranscriptRequired = errors.New("audio transcript is required")
)

// AudioMessageSender 定义音频转写后复用的消息发送能力。
type AudioMessageSender interface {
	SendMessage(input SendMessageInput) (SendMessageResult, error)
}

// AudioService 封装单段音频上传、转写和消息链路复用流程。
type AudioService struct {
	messageSender AudioMessageSender
	asr           agent.ASRClient
	maxBytes      int
}

type AudioOption func(*AudioService)

// WithMaxAudioUploadBytes 覆盖单段音频上传大小上限。
func WithMaxAudioUploadBytes(maxBytes int) AudioOption {
	return func(service *AudioService) {
		if maxBytes > 0 {
			service.maxBytes = maxBytes
		}
	}
}

// NewAudioService 创建音频上传业务服务。
func NewAudioService(messageSender AudioMessageSender, asr agent.ASRClient, opts ...AudioOption) *AudioService {
	service := &AudioService{
		messageSender: messageSender,
		asr:           asr,
		maxBytes:      MaxAudioUploadBytes,
	}
	if service.asr == nil {
		service.asr = agent.NewMockASRClient()
	}
	for _, opt := range opts {
		opt(service)
	}

	return service
}

// UploadAudioInput 是单段音频上传业务输入。
type UploadAudioInput struct {
	SessionID   int
	Filename    string
	ContentType string
	Audio       []byte
	Context     context.Context
}

// UploadAudioResult 是音频转写并进入训练链路后的业务输出。
type UploadAudioResult struct {
	Transcript string
	SendMessageResult
}

// UploadAudio 校验音频文件，调用 ASR，并将转写文本送入现有消息训练链路。
func (s *AudioService) UploadAudio(input UploadAudioInput) (UploadAudioResult, error) {
	if input.SessionID <= 0 {
		return UploadAudioResult{}, ErrInvalidAudioRequest
	}
	if len(input.Audio) == 0 {
		return UploadAudioResult{}, ErrAudioFileRequired
	}
	if len(input.Audio) > s.maxBytes {
		return UploadAudioResult{}, ErrAudioFileTooLarge
	}
	if !isSupportedAudioContentType(input.ContentType) {
		return UploadAudioResult{}, ErrAudioFileTypeUnsupported
	}

	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	asrClient := s.asr
	if asrClient == nil {
		asrClient = agent.NewMockASRClient()
	}
	output, err := asrClient.Transcribe(ctx, agent.ASRInput{
		Filename:    input.Filename,
		ContentType: input.ContentType,
		Audio:       input.Audio,
	})
	if err != nil {
		return UploadAudioResult{}, fmt.Errorf("%w: %v", ErrASRClientFailed, err)
	}

	transcript := strings.TrimSpace(output.Transcript)
	if transcript == "" {
		return UploadAudioResult{}, ErrAudioTranscriptRequired
	}
	if s.messageSender == nil {
		return UploadAudioResult{}, ErrInvalidAudioRequest
	}

	result, err := s.messageSender.SendMessage(SendMessageInput{
		SessionID: input.SessionID,
		Content:   transcript,
		Context:   ctx,
	})
	if err != nil {
		return UploadAudioResult{}, err
	}

	return UploadAudioResult{
		Transcript:        transcript,
		SendMessageResult: result,
	}, nil
}

func isSupportedAudioContentType(contentType string) bool {
	normalized := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch normalized {
	case "audio/webm", "audio/wav", "audio/wave", "audio/x-wav", "audio/mpeg", "audio/mp3", "audio/mp4", "audio/ogg", "audio/x-m4a":
		return true
	default:
		return false
	}
}
