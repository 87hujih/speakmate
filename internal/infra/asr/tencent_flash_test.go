package asr

import (
	"context"
	"errors"
	"testing"

	tencentasr "github.com/tencentcloud/tencentcloud-speech-sdk-go/asr"

	"speakmate/internal/agent"
	"speakmate/internal/config"
)

func TestNewTencentFlashClientRejectsMissingRequiredConfig(t *testing.T) {
	_, err := NewTencentFlashClient(config.ASRConfig{
		TencentAppID:      "1250000000",
		TencentSecretID:   "secret-id",
		TencentEngineType: "16k_en",
	})

	if !errors.Is(err, ErrTencentASRConfigRequired) {
		t.Fatalf("error = %v, want ErrTencentASRConfigRequired", err)
	}
}

func TestTencentFlashClientBuildsRequestAndReturnsTranscript(t *testing.T) {
	recognizer := &fakeFlashRecognizer{
		response: &tencentasr.FlashRecognitionResponse{
			RequestId: "request-1",
			Code:      0,
			FlashResult: []*tencentasr.FlashRecognitionResult{
				{Text: " I built a robot. "},
				{Text: " It used Go. "},
			},
		},
	}
	client, err := NewTencentFlashClient(validTencentASRConfig(), WithFlashRecognizer(recognizer))
	if err != nil {
		t.Fatalf("NewTencentFlashClient returned error: %v", err)
	}

	output, err := client.Transcribe(context.Background(), agent.ASRInput{
		Filename:    "answer.ogg",
		ContentType: "audio/ogg;codecs=opus",
		Audio:       []byte{0x01, 0x02},
	})
	if err != nil {
		t.Fatalf("Transcribe returned error: %v", err)
	}

	if recognizer.callCount != 1 {
		t.Fatalf("recognizer call count = %d, want 1", recognizer.callCount)
	}
	if recognizer.audio[0] != 0x01 || recognizer.audio[1] != 0x02 {
		t.Fatalf("recognizer audio = %#v, want original bytes", recognizer.audio)
	}
	if recognizer.request.EngineType != "16k_en" {
		t.Fatalf("EngineType = %q, want 16k_en", recognizer.request.EngineType)
	}
	if recognizer.request.VoiceFormat != "ogg-opus" {
		t.Fatalf("VoiceFormat = %q, want ogg-opus", recognizer.request.VoiceFormat)
	}
	if recognizer.request.HotwordId != "hotword-id" {
		t.Fatalf("HotwordId = %q, want hotword-id", recognizer.request.HotwordId)
	}
	if recognizer.request.HotwordList != "cloud word" {
		t.Fatalf("HotwordList = %q, want cloud word", recognizer.request.HotwordList)
	}
	if recognizer.request.CustomizationId != "custom-id" {
		t.Fatalf("CustomizationId = %q, want custom-id", recognizer.request.CustomizationId)
	}
	if recognizer.request.FilterDirty != 1 || recognizer.request.FilterModal != 1 || recognizer.request.FilterPunc != 1 {
		t.Fatalf("filters = %d/%d/%d, want 1/1/1", recognizer.request.FilterDirty, recognizer.request.FilterModal, recognizer.request.FilterPunc)
	}
	if recognizer.request.ConvertNumMode != 0 {
		t.Fatalf("ConvertNumMode = %d, want 0", recognizer.request.ConvertNumMode)
	}
	if recognizer.request.WordInfo != 2 {
		t.Fatalf("WordInfo = %d, want 2", recognizer.request.WordInfo)
	}
	if output.Transcript != "I built a robot. It used Go." {
		t.Fatalf("Transcript = %q, want aggregated transcript", output.Transcript)
	}
	if output.Raw != recognizer.response {
		t.Fatalf("Raw = %#v, want original response pointer", output.Raw)
	}
}

func TestTencentFlashClientMapsSupportedAudioFormats(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantFormat  string
	}{
		{name: "ogg opus", contentType: "audio/ogg", wantFormat: "ogg-opus"},
		{name: "mp4", contentType: "audio/mp4", wantFormat: "m4a"},
		{name: "x m4a", contentType: "audio/x-m4a", wantFormat: "m4a"},
		{name: "wav", contentType: "audio/wav", wantFormat: "wav"},
		{name: "x wav", contentType: "audio/x-wav", wantFormat: "wav"},
		{name: "mp3", contentType: "audio/mpeg", wantFormat: "mp3"},
		{name: "configured fallback", contentType: "", wantFormat: "ogg-opus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recognizer := &fakeFlashRecognizer{
				response: &tencentasr.FlashRecognitionResponse{
					FlashResult: []*tencentasr.FlashRecognitionResult{{Text: "hello"}},
				},
			}
			client, err := NewTencentFlashClient(validTencentASRConfig(), WithFlashRecognizer(recognizer))
			if err != nil {
				t.Fatalf("NewTencentFlashClient returned error: %v", err)
			}

			_, err = client.Transcribe(context.Background(), agent.ASRInput{
				ContentType: tt.contentType,
				Audio:       []byte{0x01},
			})
			if err != nil {
				t.Fatalf("Transcribe returned error: %v", err)
			}

			if recognizer.request.VoiceFormat != tt.wantFormat {
				t.Fatalf("VoiceFormat = %q, want %q", recognizer.request.VoiceFormat, tt.wantFormat)
			}
		})
	}
}

func TestTencentFlashClientRejectsUnsupportedWebM(t *testing.T) {
	recognizer := &fakeFlashRecognizer{}
	client, err := NewTencentFlashClient(validTencentASRConfig(), WithFlashRecognizer(recognizer))
	if err != nil {
		t.Fatalf("NewTencentFlashClient returned error: %v", err)
	}

	_, err = client.Transcribe(context.Background(), agent.ASRInput{
		ContentType: "audio/webm;codecs=opus",
		Audio:       []byte{0x01},
	})

	if !errors.Is(err, ErrTencentASRFormatUnsupported) {
		t.Fatalf("error = %v, want ErrTencentASRFormatUnsupported", err)
	}
	if recognizer.callCount != 0 {
		t.Fatalf("recognizer call count = %d, want 0", recognizer.callCount)
	}
}

func TestTencentFlashClientReturnsErrorsForEmptyAudioSDKFailureAndEmptyTranscript(t *testing.T) {
	tests := []struct {
		name       string
		input      agent.ASRInput
		response   *tencentasr.FlashRecognitionResponse
		sdkErr     error
		wantErr    error
		wantCalled bool
	}{
		{
			name:    "empty audio",
			input:   agent.ASRInput{ContentType: "audio/ogg"},
			wantErr: agent.ErrASRAudioRequired,
		},
		{
			name:       "sdk error",
			input:      agent.ASRInput{ContentType: "audio/ogg", Audio: []byte{0x01}},
			sdkErr:     errors.New("network unavailable"),
			wantErr:    ErrTencentASRRecognizeFailed,
			wantCalled: true,
		},
		{
			name:       "tencent response error",
			input:      agent.ASRInput{ContentType: "audio/ogg", Audio: []byte{0x01}},
			response:   &tencentasr.FlashRecognitionResponse{Code: 100, Message: "auth failed"},
			wantErr:    ErrTencentASRRecognizeFailed,
			wantCalled: true,
		},
		{
			name:       "empty transcript",
			input:      agent.ASRInput{ContentType: "audio/ogg", Audio: []byte{0x01}},
			response:   &tencentasr.FlashRecognitionResponse{FlashResult: []*tencentasr.FlashRecognitionResult{{Text: "   "}}},
			wantErr:    ErrTencentASREmptyTranscript,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recognizer := &fakeFlashRecognizer{response: tt.response, err: tt.sdkErr}
			client, err := NewTencentFlashClient(validTencentASRConfig(), WithFlashRecognizer(recognizer))
			if err != nil {
				t.Fatalf("NewTencentFlashClient returned error: %v", err)
			}

			_, err = client.Transcribe(context.Background(), tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if gotCalled := recognizer.callCount > 0; gotCalled != tt.wantCalled {
				t.Fatalf("recognizer called = %v, want %v", gotCalled, tt.wantCalled)
			}
		})
	}
}

type fakeFlashRecognizer struct {
	request   *tencentasr.FlashRecognitionRequest
	audio     []byte
	response  *tencentasr.FlashRecognitionResponse
	err       error
	callCount int
}

func (r *fakeFlashRecognizer) Recognize(req *tencentasr.FlashRecognitionRequest, audio []byte) (*tencentasr.FlashRecognitionResponse, error) {
	r.callCount++
	r.request = req
	r.audio = append([]byte(nil), audio...)
	if r.err != nil {
		return nil, r.err
	}

	return r.response, nil
}

func validTencentASRConfig() config.ASRConfig {
	return config.ASRConfig{
		TencentAppID:           "1250000000",
		TencentSecretID:        "secret-id",
		TencentSecretKey:       "secret-key",
		TencentEngineType:      "16k_en",
		TencentVoiceFormat:     "ogg-opus",
		TencentHotwordID:       "hotword-id",
		TencentHotwordList:     "cloud word",
		TencentCustomizationID: "custom-id",
		TencentFilterDirty:     1,
		TencentFilterModal:     1,
		TencentFilterPunc:      1,
		TencentConvertNumMode:  0,
		TencentWordInfo:        2,
	}
}
