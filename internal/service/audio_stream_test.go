package service_test

import (
	"context"
	"errors"
	"testing"

	"speakmate/internal/agent"
	"speakmate/internal/service"
)

func TestAudioStreamSkipsASROnAppendChunkWhenRealtimePartialsDisabled(t *testing.T) {
	asr := &fakeASRClient{
		output: agent.ASROutput{Transcript: "I built a speech practice app."},
	}
	messageSender := &fakeAudioMessageSender{
		result: sampleAudioMessageResult("I built a speech practice app."),
	}
	streamService := service.NewAudioStreamService(
		messageSender,
		asr,
		service.WithAudioStreamPartialTranscription(false),
	)
	stream, err := streamService.Start(service.StartAudioStreamInput{
		SessionID:   7,
		ContentType: "audio/ogg",
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	first, err := stream.AppendChunk(service.AudioStreamChunkInput{
		Audio:    []byte{0x01, 0x02},
		Sequence: 3,
	})
	if err != nil {
		t.Fatalf("AppendChunk first returned error: %v", err)
	}
	second, err := stream.AppendChunk(service.AudioStreamChunkInput{
		Audio: []byte{0x03},
	})
	if err != nil {
		t.Fatalf("AppendChunk second returned error: %v", err)
	}

	if first.Transcript != "" || second.Transcript != "" {
		t.Fatalf("partials = %q/%q, want empty placeholders", first.Transcript, second.Transcript)
	}
	if first.Sequence != 3 {
		t.Fatalf("first sequence = %d, want explicit sequence 3", first.Sequence)
	}
	if second.Sequence != 2 {
		t.Fatalf("second sequence = %d, want chunk count 2", second.Sequence)
	}
	if asr.callCount != 0 {
		t.Fatalf("asr call count after append = %d, want 0", asr.callCount)
	}

	result, err := stream.Finish(context.Background())
	if err != nil {
		t.Fatalf("Finish returned error: %v", err)
	}

	if asr.callCount != 1 {
		t.Fatalf("asr call count after finish = %d, want 1", asr.callCount)
	}
	if asr.input.ContentType != "audio/ogg" {
		t.Fatalf("asr content type = %q, want audio/ogg", asr.input.ContentType)
	}
	if string(asr.input.Audio) != string([]byte{0x01, 0x02, 0x03}) {
		t.Fatalf("asr audio = %#v, want concatenated chunks", asr.input.Audio)
	}
	if result.Transcript != "I built a speech practice app." {
		t.Fatalf("final transcript = %q, want ASR transcript", result.Transcript)
	}
	if messageSender.callCount != 1 {
		t.Fatalf("message sender call count = %d, want 1", messageSender.callCount)
	}
}

func TestAudioStreamKeepsMockPartialASRByDefault(t *testing.T) {
	asr := &fakeASRClient{
		output: agent.ASROutput{Transcript: "I am study computer science and I have did a project."},
	}
	messageSender := &fakeAudioMessageSender{
		result: sampleAudioMessageResult("I am study computer science and I have did a project."),
	}
	streamService := service.NewAudioStreamService(messageSender, asr)
	stream, err := streamService.Start(service.StartAudioStreamInput{
		SessionID:   7,
		ContentType: "audio/webm",
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	partial, err := stream.AppendChunk(service.AudioStreamChunkInput{
		Audio:    []byte{0x01},
		Sequence: 1,
	})
	if err != nil {
		t.Fatalf("AppendChunk returned error: %v", err)
	}
	if partial.Transcript == "" {
		t.Fatal("partial transcript is empty, want mock partial")
	}
	if asr.callCount != 1 {
		t.Fatalf("asr call count after append = %d, want 1", asr.callCount)
	}

	_, err = stream.Finish(context.Background())
	if err != nil {
		t.Fatalf("Finish returned error: %v", err)
	}
	if asr.callCount != 2 {
		t.Fatalf("asr call count after finish = %d, want 2", asr.callCount)
	}
}

func TestAudioStreamFinishReturnsASRFailureWithoutSendingMessage(t *testing.T) {
	asr := &fakeASRClient{err: errors.New("asr failed")}
	messageSender := &fakeAudioMessageSender{}
	streamService := service.NewAudioStreamService(
		messageSender,
		asr,
		service.WithAudioStreamPartialTranscription(false),
	)
	stream, err := streamService.Start(service.StartAudioStreamInput{
		SessionID:   7,
		ContentType: "audio/ogg",
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if _, err := stream.AppendChunk(service.AudioStreamChunkInput{Audio: []byte{0x01}}); err != nil {
		t.Fatalf("AppendChunk returned error: %v", err)
	}

	_, err = stream.Finish(context.Background())

	if !errors.Is(err, service.ErrASRClientFailed) {
		t.Fatalf("error = %v, want ErrASRClientFailed", err)
	}
	if messageSender.callCount != 0 {
		t.Fatalf("message sender call count = %d, want 0", messageSender.callCount)
	}
}
