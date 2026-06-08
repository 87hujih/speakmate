package main

import (
	"log"
	"strings"

	"speakmate/internal/config"
	"speakmate/internal/router"
	"speakmate/internal/security"
)

func main() {
	cfg := config.Load()
	log.Printf(
		"starting SpeakMate backend: asr_provider=%s asr_use_mock=%t asr_mock_transcript_configured=%t",
		cfg.ASR.Provider,
		cfg.ASR.UseMock,
		strings.TrimSpace(cfg.ASR.MockTranscript) != "",
	)
	engine, err := router.NewWithError(cfg)
	if err != nil {
		log.Fatalf("init server failed: %s", security.RedactString(err.Error()))
	}

	if err := engine.Run(cfg.Addr()); err != nil {
		log.Fatalf("server stopped: %s", security.RedactString(err.Error()))
	}
}
