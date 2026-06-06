package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port string
	LLM  LLMConfig
}

type LLMConfig struct {
	Provider       string
	BaseURL        string
	APIKey         string
	Model          string
	TimeoutSeconds int
	UseMock        bool
}

func Load() Config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		Port: port,
		LLM: LLMConfig{
			Provider:       stringEnv("LLM_PROVIDER", "openai-compatible"),
			BaseURL:        strings.TrimSpace(os.Getenv("LLM_BASE_URL")),
			APIKey:         strings.TrimSpace(os.Getenv("LLM_API_KEY")),
			Model:          strings.TrimSpace(os.Getenv("LLM_MODEL")),
			TimeoutSeconds: positiveIntEnv("LLM_TIMEOUT_SECONDS", 30),
			UseMock:        boolEnv("LLM_USE_MOCK", true),
		},
	}
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func (c LLMConfig) HasRequiredFields() bool {
	return strings.TrimSpace(c.BaseURL) != "" &&
		strings.TrimSpace(c.APIKey) != "" &&
		strings.TrimSpace(c.Model) != ""
}

func stringEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func positiveIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func boolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
