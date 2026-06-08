package agent

import (
	"context"
	"errors"
	"strings"
)

var ErrMockASRTranscriptRequired = errors.New("mock asr transcript required")

// MockASRClient 返回稳定转写文本，供本地开发和自动测试使用。
type MockASRClient struct {
	transcript string
}

type MockASROption func(*MockASRClient)

// WithMockASRTranscript 覆盖 Mock ASR 的稳定转写文本。
func WithMockASRTranscript(transcript string) MockASROption {
	return func(client *MockASRClient) {
		client.transcript = transcript
	}
}

// NewMockASRClient 创建 Mock ASR Client。
func NewMockASRClient(opts ...MockASROption) *MockASRClient {
	client := &MockASRClient{}
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Transcribe 返回确定性的 Mock 转写结果。
func (c *MockASRClient) Transcribe(ctx context.Context, input ASRInput) (ASROutput, error) {
	if len(input.Audio) == 0 {
		return ASROutput{}, ErrASRAudioRequired
	}

	transcript := strings.TrimSpace(c.transcript)
	if transcript == "" {
		return ASROutput{}, ErrMockASRTranscriptRequired
	}

	return ASROutput{
		Transcript: transcript,
		Confidence: 0.92,
		Raw:        nil,
	}, nil
}
