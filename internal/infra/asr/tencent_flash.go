package asr

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tencentasr "github.com/tencentcloud/tencentcloud-speech-sdk-go/asr"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/common"

	"speakmate/internal/agent"
	"speakmate/internal/config"
)

// 基础设施层复用的哨兵错误。
var (
	ErrTencentASRConfigRequired    = errors.New("腾讯 ASR 配置不能为空")
	ErrTencentASRFormatUnsupported = errors.New("不支持的腾讯 ASR 音频格式")
	ErrTencentASRRecognizeFailed   = errors.New("腾讯 ASR 识别失败")
	ErrTencentASREmptyTranscript   = errors.New("腾讯 ASR 转写结果为空")
)

// flashRecognizer 抽象腾讯极速 ASR SDK 调用，便于测试替换。
type flashRecognizer interface {
	Recognize(req *tencentasr.FlashRecognitionRequest, audio []byte) (*tencentasr.FlashRecognitionResponse, error)
}

// TencentFlashClient 封装腾讯极速 ASR 单段识别能力。
type TencentFlashClient struct {
	cfg        config.ASRConfig
	recognizer flashRecognizer
}

// TencentFlashOption 用于配置 TencentFlashClient。
type TencentFlashOption func(*TencentFlashClient)

// WithFlashRecognizer 返回用于覆盖默认行为的配置选项。
func WithFlashRecognizer(recognizer flashRecognizer) TencentFlashOption {
	return func(client *TencentFlashClient) {
		if recognizer != nil {
			client.recognizer = recognizer
		}
	}
}

// NewTencentFlashClient 创建并返回对应组件实例。
func NewTencentFlashClient(cfg config.ASRConfig, opts ...TencentFlashOption) (*TencentFlashClient, error) {
	if !cfg.HasTencentRequiredFields() {
		return nil, ErrTencentASRConfigRequired
	}

	client := &TencentFlashClient{
		cfg: cfg,
		recognizer: tencentasr.NewFlashRecognizer(
			strings.TrimSpace(cfg.TencentAppID),
			common.NewCredential(strings.TrimSpace(cfg.TencentSecretID), strings.TrimSpace(cfg.TencentSecretKey)),
		),
	}
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Transcribe 调用腾讯极速 ASR 完成单段音频转写。
func (c *TencentFlashClient) Transcribe(ctx context.Context, input agent.ASRInput) (agent.ASROutput, error) {
	if len(input.Audio) == 0 {
		return agent.ASROutput{}, agent.ErrASRAudioRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return agent.ASROutput{}, fmt.Errorf("%w: %v", ErrTencentASRRecognizeFailed, err)
	}

	voiceFormat, err := c.voiceFormat(input.ContentType)
	if err != nil {
		return agent.ASROutput{}, err
	}

	req := &tencentasr.FlashRecognitionRequest{
		EngineType:      strings.TrimSpace(c.cfg.TencentEngineType),
		VoiceFormat:     voiceFormat,
		HotwordId:       strings.TrimSpace(c.cfg.TencentHotwordID),
		HotwordList:     strings.TrimSpace(c.cfg.TencentHotwordList),
		CustomizationId: strings.TrimSpace(c.cfg.TencentCustomizationID),
		FilterDirty:     int32(c.cfg.TencentFilterDirty),
		FilterModal:     int32(c.cfg.TencentFilterModal),
		FilterPunc:      int32(c.cfg.TencentFilterPunc),
		ConvertNumMode:  int32(c.cfg.TencentConvertNumMode),
		WordInfo:        int32(c.cfg.TencentWordInfo),
	}

	response, err := c.recognizer.Recognize(req, input.Audio)
	if err != nil {
		return agent.ASROutput{}, fmt.Errorf("%w: %v", ErrTencentASRRecognizeFailed, err)
	}
	if response == nil {
		return agent.ASROutput{}, fmt.Errorf("%w：响应为空", ErrTencentASRRecognizeFailed)
	}
	if response.Code != 0 {
		return agent.ASROutput{}, fmt.Errorf("%w：错误码 %d：%s", ErrTencentASRRecognizeFailed, response.Code, response.Message)
	}

	transcript := aggregateTranscript(response.FlashResult)
	if transcript == "" {
		return agent.ASROutput{}, ErrTencentASREmptyTranscript
	}

	return agent.ASROutput{
		Transcript: transcript,
		Raw:        response,
	}, nil
}

// voiceFormat 根据音频 MIME 类型推导腾讯 ASR voice_format。
func (c *TencentFlashClient) voiceFormat(contentType string) (string, error) {
	normalized := normalizeContentType(contentType)
	switch normalized {
	case "":
		fallback := strings.TrimSpace(c.cfg.TencentVoiceFormat)
		if fallback == "" {
			fallback = "ogg-opus"
		}
		return fallback, nil
	case "audio/ogg":
		return "ogg-opus", nil
	case "audio/mp4", "audio/m4a", "audio/x-m4a":
		return "m4a", nil
	case "audio/wav", "audio/wave", "audio/x-wav":
		return "wav", nil
	case "audio/mpeg", "audio/mp3":
		return "mp3", nil
	case "audio/aac":
		return "aac", nil
	case "audio/amr":
		return "amr", nil
	case "audio/pcm":
		return "pcm", nil
	default:
		return "", fmt.Errorf("%w: %s", ErrTencentASRFormatUnsupported, normalized)
	}
}

// normalizeContentType 归一化上传音频的 Content-Type。
func normalizeContentType(contentType string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

// aggregateTranscript 从腾讯 ASR 响应中合并识别文本。
func aggregateTranscript(results []*tencentasr.FlashRecognitionResult) string {
	parts := make([]string, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		text := strings.TrimSpace(result.Text)
		if text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " ")
}
