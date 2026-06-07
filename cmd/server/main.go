package main

import (
	"log"

	"speakmate/internal/config"
	"speakmate/internal/router"
)

func main() {
	cfg := config.Load()
	engine, err := router.NewWithError(cfg)
	if err != nil {
		log.Fatalf("init server failed: %v", err)
	}

	if err := engine.Run(cfg.Addr()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
