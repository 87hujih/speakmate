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

var (
	ErrTencentASRConfigRequired    = errors.New("tencent asr config required")
	ErrTencentASRFormatUnsupported = errors.New("tencent asr audio format unsupported")
	ErrTencentASRRecognizeFailed   = errors.New("tencent asr recognize failed")
	ErrTencentASREmptyTranscript   = errors.New("tencent asr transcript empty")
)

type flashRecognizer interface {
	Recognize(req *tencentasr.FlashRecognitionRequest, audio []byte) (*tencentasr.FlashRecognitionResponse, error)
}

type TencentFlashClient struct {
	cfg        config.ASRConfig
	recognizer flashRecognizer
}

type TencentFlashOption func(*TencentFlashClient)

func WithFlashRecognizer(recognizer flashRecognizer) TencentFlashOption {
	return func(client *TencentFlashClient) {
		if recognizer != nil {
			client.recognizer = recognizer
		}
	}
}

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
		return agent.ASROutput{}, fmt.Errorf("%w: empty response", ErrTencentASRRecognizeFailed)
	}
	if response.Code != 0 {
		return agent.ASROutput{}, fmt.Errorf("%w: code %d: %s", ErrTencentASRRecognizeFailed, response.Code, response.Message)
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

func normalizeContentType(contentType string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

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
