package main

import (
	"log"
	"strings"

	"speakmate/internal/config"
	"speakmate/internal/router"
	"speakmate/internal/security"
)

// main 是当前命令的入口，负责串联配置加载和执行流程。
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
		log.Fatalf("服务初始化失败：%s", security.RedactString(err.Error()))
	}

	if err := engine.Run(cfg.Addr()); err != nil {
		log.Fatalf("服务停止：%s", security.RedactString(err.Error()))
	}
}
