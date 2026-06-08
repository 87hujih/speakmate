package agent

import (
	"context"
	"errors"
	"testing"
)

func TestMockASRClientRejectsMissingConfiguredTranscript(t *testing.T) {
	client := NewMockASRClient()

	_, err := client.Transcribe(context.Background(), ASRInput{
		Filename:    "answer.webm",
		ContentType: "audio/webm",
		Audio:       []byte{0x01, 0x02, 0x03},
	})

	if !errors.Is(err, ErrMockASRTranscriptRequired) {
		t.Fatalf("error = %v, want ErrMockASRTranscriptRequired", err)
	}
}

func TestMockASRClientTrimsConfiguredTranscript(t *testing.T) {
	client := NewMockASRClient(WithMockASRTranscript("  Could you recommend something light?  "))

	output, err := client.Transcribe(context.Background(), ASRInput{
		Filename:    "answer.wav",
		ContentType: "audio/wav",
		Audio:       []byte{0x01},
	})
	if err != nil {
		t.Fatalf("Transcribe returned error: %v", err)
	}

	if output.Transcript != "Could you recommend something light?" {
		t.Fatalf("transcript = %q, want trimmed configured transcript", output.Transcript)
	}
	if output.Confidence <= 0 {
		t.Fatalf("confidence = %f, want positive", output.Confidence)
	}
}

func TestMockASRClientRejectsEmptyAudio(t *testing.T) {
	client := NewMockASRClient()

	_, err := client.Transcribe(context.Background(), ASRInput{
		Filename:    "empty.webm",
		ContentType: "audio/webm",
		Audio:       []byte{},
	})

	if !errors.Is(err, ErrASRAudioRequired) {
		t.Fatalf("error = %v, want ErrASRAudioRequired", err)
	}
}
