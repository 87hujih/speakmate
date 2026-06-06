package config

import "testing"

func TestLoadDefaultsToPort8080(t *testing.T) {
	t.Setenv("APP_PORT", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "8080")
	}

	if cfg.Addr() != ":8080" {
		t.Fatalf("Addr() = %q, want %q", cfg.Addr(), ":8080")
	}
}
