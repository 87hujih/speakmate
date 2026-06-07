package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"speakmate/internal/agent"
	"speakmate/internal/model"
	"speakmate/internal/service"
)

func TestAudioServiceTranscribesAudioAndSendsTranscriptThroughMessageFlow(t *testing.T) {
	asr := &fakeASRClient{
		output: agent.ASROutput{
			Transcript: " I built a robot control project. ",
			Confidence: 0.93,
		},
	}
	messageSender := &fakeAudioMessageSender{
		result: sampleAudioMessageResult("I built a robot control project."),
	}
	audioService := service.NewAudioService(messageSender, asr)

	result, err := audioService.UploadAudio(service.UploadAudioInput{
		SessionID:   7,
		Filename:    "answer.webm",
		ContentType: "audio/webm",
		Audio:       []byte{0x01, 0x02},
		Context:     context.Background(),
	})
	if err != nil {
		t.Fatalf("UploadAudio returned error: %v", err)
	}

	if asr.callCount != 1 {
		t.Fatalf("asr call count = %d, want 1", asr.callCount)
	}
	if asr.input.Filename != "answer.webm" {
		t.Fatalf("asr filename = %q, want answer.webm", asr.input.Filename)
	}
	if asr.input.ContentType != "audio/webm" {
		t.Fatalf("asr content type = %q, want audio/webm", asr.input.ContentType)
	}
	if messageSender.callCount != 1 {
		t.Fatalf("message sender call count = %d, want 1", messageSender.callCount)
	}
	if messageSender.input.SessionID != 7 {
		t.Fatalf("message sender session id = %d, want 7", messageSender.input.SessionID)
	}
	if messageSender.input.Content != "I built a robot control project." {
		t.Fatalf("message sender content = %q, want transcript", messageSender.input.Content)
	}
	if result.Transcript != "I built a robot control project." {
		t.Fatalf("result transcript = %q, want trimmed transcript", result.Transcript)
	}
	if result.UserMessage.Content != "I built a robot control project." {
		t.Fatalf("user message content = %q, want transcript", result.UserMessage.Content)
	}
}

func TestAudioServiceRejectsInvalidAudioInputBeforeASR(t *testing.T) {
	tests := []struct {
		name  string
		input service.UploadAudioInput
		want  error
	}{
		{
			name: "invalid session id",
			input: service.UploadAudioInput{
				SessionID:   0,
				Filename:    "answer.webm",
				ContentType: "audio/webm",
				Audio:       []byte{0x01},
			},
			want: service.ErrInvalidAudioRequest,
		},
		{
			name: "empty audio",
			input: service.UploadAudioInput{
				SessionID:   7,
				Filename:    "answer.webm",
				ContentType: "audio/webm",
				Audio:       []byte{},
			},
			want: service.ErrAudioFileRequired,
		},
		{
			name: "unsupported content type",
			input: service.UploadAudioInput{
				SessionID:   7,
				Filename:    "answer.txt",
				ContentType: "text/plain",
				Audio:       []byte{0x01},
			},
			want: service.ErrAudioFileTypeUnsupported,
		},
		{
			name: "file too large",
			input: service.UploadAudioInput{
				SessionID:   7,
				Filename:    "answer.webm",
				ContentType: "audio/webm",
				Audio:       []byte{0x01, 0x02, 0x03},
			},
			want: service.ErrAudioFileTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asr := &fakeASRClient{}
			messageSender := &fakeAudioMessageSender{}
			audioService := service.NewAudioService(messageSender, asr, service.WithMaxAudioUploadBytes(2))

			_, err := audioService.UploadAudio(tt.input)

			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if asr.callCount != 0 {
				t.Fatalf("asr call count = %d, want 0", asr.callCount)
			}
			if messageSender.callCount != 0 {
				t.Fatalf("message sender call count = %d, want 0", messageSender.callCount)
			}
		})
	}
}

func TestAudioServiceReturnsASRFailureWithoutSendingMessage(t *testing.T) {
	asr := &fakeASRClient{err: errors.New("asr unavailable")}
	messageSender := &fakeAudioMessageSender{}
	audioService := service.NewAudioService(messageSender, asr)

	_, err := audioService.UploadAudio(service.UploadAudioInput{
		SessionID:   7,
		Filename:    "answer.webm",
		ContentType: "audio/webm",
		Audio:       []byte{0x01},
	})

	if !errors.Is(err, service.ErrASRClientFailed) {
		t.Fatalf("error = %v, want ErrASRClientFailed", err)
	}
	if messageSender.callCount != 0 {
		t.Fatalf("message sender call count = %d, want 0", messageSender.callCount)
	}
}

func TestAudioServiceRejectsBlankTranscript(t *testing.T) {
	asr := &fakeASRClient{output: agent.ASROutput{Transcript: "   "}}
	messageSender := &fakeAudioMessageSender{}
	audioService := service.NewAudioService(messageSender, asr)

	_, err := audioService.UploadAudio(service.UploadAudioInput{
		SessionID:   7,
		Filename:    "answer.webm",
		ContentType: "audio/webm",
		Audio:       []byte{0x01},
	})

	if !errors.Is(err, service.ErrAudioTranscriptRequired) {
		t.Fatalf("error = %v, want ErrAudioTranscriptRequired", err)
	}
	if messageSender.callCount != 0 {
		t.Fatalf("message sender call count = %d, want 0", messageSender.callCount)
	}
}

func TestAudioServicePropagatesFinishedSessionErrorFromMessageFlow(t *testing.T) {
	asr := &fakeASRClient{output: agent.ASROutput{Transcript: "Can we continue?"}}
	messageSender := &fakeAudioMessageSender{err: service.ErrSessionAlreadyFinished}
	audioService := service.NewAudioService(messageSender, asr)

	_, err := audioService.UploadAudio(service.UploadAudioInput{
		SessionID:   7,
		Filename:    "answer.webm",
		ContentType: "audio/webm",
		Audio:       []byte{0x01},
	})

	if !errors.Is(err, service.ErrSessionAlreadyFinished) {
		t.Fatalf("error = %v, want ErrSessionAlreadyFinished", err)
	}
	if messageSender.callCount != 1 {
		t.Fatalf("message sender call count = %d, want 1", messageSender.callCount)
	}
}

type fakeASRClient struct {
	output    agent.ASROutput
	err       error
	callCount int
	input     agent.ASRInput
}

func (c *fakeASRClient) Transcribe(ctx context.Context, input agent.ASRInput) (agent.ASROutput, error) {
	c.callCount++
	c.input = input
	if c.err != nil {
		return agent.ASROutput{}, c.err
	}

	return c.output, nil
}

type fakeAudioMessageSender struct {
	result    service.SendMessageResult
	err       error
	callCount int
	input     service.SendMessageInput
}

func (s *fakeAudioMessageSender) SendMessage(input service.SendMessageInput) (service.SendMessageResult, error) {
	s.callCount++
	s.input = input
	if s.err != nil {
		return service.SendMessageResult{}, s.err
	}

	return s.result, nil
}

func sampleAudioMessageResult(content string) service.SendMessageResult {
	createdAt := time.Date(2026, 6, 7, 3, 0, 0, 0, time.UTC)

	return service.SendMessageResult{
		UserMessage: model.Message{
			ID:        11,
			SessionID: 7,
			Role:      model.MessageRoleUser,
			Content:   content,
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
		TurnCount: 2,
		CorrectionSummary: service.CorrectionSummary{
			HasErrors:  false,
			ErrorCount: 0,
		},
		ScoreSummary: service.ScoreSummary{
			TotalScore: 86,
			Grammar:    88,
			Expression: 84,
		},
	}
}
