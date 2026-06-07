package main

import (
	"log"

	"speakmate/internal/config"
	"speakmate/internal/router"
	"speakmate/internal/security"
)

func main() {
	cfg := config.Load()
	engine, err := router.NewWithError(cfg)
	if err != nil {
		log.Fatalf("init server failed: %s", security.RedactString(err.Error()))
	}

	if err := engine.Run(cfg.Addr()); err != nil {
		log.Fatalf("server stopped: %s", security.RedactString(err.Error()))
	}
}
