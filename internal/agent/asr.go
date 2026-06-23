package agent

import (
	"context"
	"errors"
)

// Agent 层复用的哨兵错误。
var (
	// ErrASRAudioRequired 表示 ASR 输入缺少音频内容。
	ErrASRAudioRequired = errors.New("ASR 音频不能为空")
)

// ASRClient 定义音频转文本能力。
type ASRClient interface {
	Transcribe(ctx context.Context, input ASRInput) (ASROutput, error)
}

// ASRInput 是一次单段音频转写请求。
type ASRInput struct {
	Filename    string
	ContentType string
	Audio       []byte
}

// ASROutput 是 ASR 转写结果。
type ASROutput struct {
	Transcript string
	Confidence float64
	Raw        any
}
