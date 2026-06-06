package main

import (
	"log"

	"speakmate/internal/config"
	"speakmate/internal/router"
)

func main() {
	cfg := config.Load()
	engine := router.New(cfg)

	if err := engine.Run(cfg.Addr()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
