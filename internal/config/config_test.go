package config

import "testing"

func TestLoadDefaultsToPort8080(t *testing.T) {
	t.Setenv("APP_PORT", "")
	clearLLMEnv(t)
	clearStorageEnv(t)

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "8080")
	}

	if cfg.Addr() != ":8080" {
		t.Fatalf("Addr() = %q, want %q", cfg.Addr(), ":8080")
	}

	if cfg.LLM.Provider != "openai-compatible" {
		t.Fatalf("LLM.Provider = %q, want openai-compatible", cfg.LLM.Provider)
	}
	if cfg.LLM.BaseURL != "" {
		t.Fatalf("LLM.BaseURL = %q, want empty", cfg.LLM.BaseURL)
	}
	if cfg.LLM.APIKey != "" {
		t.Fatalf("LLM.APIKey = %q, want empty", cfg.LLM.APIKey)
	}
	if cfg.LLM.Model != "" {
		t.Fatalf("LLM.Model = %q, want empty", cfg.LLM.Model)
	}
	if cfg.LLM.TimeoutSeconds != 30 {
		t.Fatalf("LLM.TimeoutSeconds = %d, want 30", cfg.LLM.TimeoutSeconds)
	}
	if !cfg.LLM.UseMock {
		t.Fatal("LLM.UseMock = false, want true by default")
	}
	if !cfg.Feedback.CorrectionUseMock {
		t.Fatal("Feedback.CorrectionUseMock = false, want true by default")
	}
	if !cfg.Feedback.ScoringUseMock {
		t.Fatal("Feedback.ScoringUseMock = false, want true by default")
	}
	if !cfg.Feedback.SummaryUseMock {
		t.Fatal("Feedback.SummaryUseMock = false, want true by default")
	}
	if !cfg.Feedback.FailOpen {
		t.Fatal("Feedback.FailOpen = false, want true by default")
	}
	if cfg.Storage.Mode != StorageModeMemory {
		t.Fatalf("Storage.Mode = %q, want %q", cfg.Storage.Mode, StorageModeMemory)
	}
	if cfg.Storage.MySQLDSN != "" {
		t.Fatalf("Storage.MySQLDSN = %q, want empty", cfg.Storage.MySQLDSN)
	}
	if err := cfg.Storage.Validate(); err != nil {
		t.Fatalf("Storage.Validate returned error for default memory mode: %v", err)
	}
}

func TestLoadReadsLLMEnvironment(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	t.Setenv("LLM_PROVIDER", "openai-compatible")
	t.Setenv("LLM_BASE_URL", "https://llm.example.com/v1")
	t.Setenv("LLM_API_KEY", "test-key")
	t.Setenv("LLM_MODEL", "test-model")
	t.Setenv("LLM_TIMEOUT_SECONDS", "45")
	t.Setenv("LLM_USE_MOCK", "false")
	t.Setenv("CORRECTION_USE_MOCK", "false")
	t.Setenv("SCORING_USE_MOCK", "false")
	t.Setenv("SUMMARY_USE_MOCK", "false")
	t.Setenv("FEEDBACK_FAIL_OPEN", "false")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.LLM.Provider != "openai-compatible" {
		t.Fatalf("LLM.Provider = %q, want openai-compatible", cfg.LLM.Provider)
	}
	if cfg.LLM.BaseURL != "https://llm.example.com/v1" {
		t.Fatalf("LLM.BaseURL = %q, want https://llm.example.com/v1", cfg.LLM.BaseURL)
	}
	if cfg.LLM.APIKey != "test-key" {
		t.Fatalf("LLM.APIKey = %q, want test-key", cfg.LLM.APIKey)
	}
	if cfg.LLM.Model != "test-model" {
		t.Fatalf("LLM.Model = %q, want test-model", cfg.LLM.Model)
	}
	if cfg.LLM.TimeoutSeconds != 45 {
		t.Fatalf("LLM.TimeoutSeconds = %d, want 45", cfg.LLM.TimeoutSeconds)
	}
	if cfg.LLM.UseMock {
		t.Fatal("LLM.UseMock = true, want false")
	}
	if cfg.Feedback.CorrectionUseMock {
		t.Fatal("Feedback.CorrectionUseMock = true, want false")
	}
	if cfg.Feedback.ScoringUseMock {
		t.Fatal("Feedback.ScoringUseMock = true, want false")
	}
	if cfg.Feedback.SummaryUseMock {
		t.Fatal("Feedback.SummaryUseMock = true, want false")
	}
	if cfg.Feedback.FailOpen {
		t.Fatal("Feedback.FailOpen = true, want false")
	}
}

func TestLoadReadsStorageEnvironment(t *testing.T) {
	clearLLMEnv(t)
	t.Setenv("STORAGE_MODE", "mysql")
	t.Setenv("MYSQL_DSN", "speakmate:secret@tcp(127.0.0.1:3306)/speakmate?parseTime=true")

	cfg := Load()

	if cfg.Storage.Mode != StorageModeMySQL {
		t.Fatalf("Storage.Mode = %q, want %q", cfg.Storage.Mode, StorageModeMySQL)
	}
	if cfg.Storage.MySQLDSN != "speakmate:secret@tcp(127.0.0.1:3306)/speakmate?parseTime=true" {
		t.Fatalf("Storage.MySQLDSN = %q, want configured DSN", cfg.Storage.MySQLDSN)
	}
	if err := cfg.Storage.Validate(); err != nil {
		t.Fatalf("Storage.Validate returned error for valid mysql mode: %v", err)
	}
}

func TestStorageValidateRejectsMissingMySQLDSN(t *testing.T) {
	cfg := StorageConfig{Mode: StorageModeMySQL}

	err := cfg.Validate()

	if err == nil {
		t.Fatal("Validate returned nil, want missing MySQL DSN error")
	}
}

func TestStorageValidateRejectsUnknownMode(t *testing.T) {
	cfg := StorageConfig{Mode: "postgres", MySQLDSN: "dsn"}

	err := cfg.Validate()

	if err == nil {
		t.Fatal("Validate returned nil, want unknown storage mode error")
	}
}

func TestLoadFallsBackForInvalidLLMTimeoutAndMockFlag(t *testing.T) {
	clearLLMEnv(t)
	t.Setenv("LLM_TIMEOUT_SECONDS", "not-a-number")
	t.Setenv("LLM_USE_MOCK", "not-a-bool")
	t.Setenv("CORRECTION_USE_MOCK", "not-a-bool")
	t.Setenv("SCORING_USE_MOCK", "not-a-bool")
	t.Setenv("SUMMARY_USE_MOCK", "not-a-bool")
	t.Setenv("FEEDBACK_FAIL_OPEN", "not-a-bool")

	cfg := Load()

	if cfg.LLM.TimeoutSeconds != 30 {
		t.Fatalf("LLM.TimeoutSeconds = %d, want 30", cfg.LLM.TimeoutSeconds)
	}
	if !cfg.LLM.UseMock {
		t.Fatal("LLM.UseMock = false, want true for invalid flag")
	}
	if !cfg.Feedback.CorrectionUseMock {
		t.Fatal("Feedback.CorrectionUseMock = false, want true for invalid flag")
	}
	if !cfg.Feedback.ScoringUseMock {
		t.Fatal("Feedback.ScoringUseMock = false, want true for invalid flag")
	}
	if !cfg.Feedback.SummaryUseMock {
		t.Fatal("Feedback.SummaryUseMock = false, want true for invalid flag")
	}
	if !cfg.Feedback.FailOpen {
		t.Fatal("Feedback.FailOpen = false, want true for invalid flag")
	}
}

func clearLLMEnv(t *testing.T) {
	t.Helper()

	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("LLM_BASE_URL", "")
	t.Setenv("LLM_API_KEY", "")
	t.Setenv("LLM_MODEL", "")
	t.Setenv("LLM_TIMEOUT_SECONDS", "")
	t.Setenv("LLM_USE_MOCK", "")
	t.Setenv("CORRECTION_USE_MOCK", "")
	t.Setenv("SCORING_USE_MOCK", "")
	t.Setenv("SUMMARY_USE_MOCK", "")
	t.Setenv("FEEDBACK_FAIL_OPEN", "")
}

func clearStorageEnv(t *testing.T) {
	t.Helper()

	t.Setenv("STORAGE_MODE", "")
	t.Setenv("MYSQL_DSN", "")
}
