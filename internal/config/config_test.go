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
	if !cfg.LLM.FallbackToMock {
		t.Fatal("LLM.FallbackToMock = false, want true by default")
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
	if cfg.ExternalServiceTimeoutSeconds != 30 {
		t.Fatalf("ExternalServiceTimeoutSeconds = %d, want 30", cfg.ExternalServiceTimeoutSeconds)
	}
	if cfg.Server.RequestTimeoutSeconds != 30 {
		t.Fatalf("Server.RequestTimeoutSeconds = %d, want 30", cfg.Server.RequestTimeoutSeconds)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Fatalf("CORS.AllowedOrigins length = %d, want 2", len(cfg.CORS.AllowedOrigins))
	}
	if cfg.CORS.AllowedOrigins[0] != "http://localhost:5173" {
		t.Fatalf("CORS.AllowedOrigins[0] = %q, want http://localhost:5173", cfg.CORS.AllowedOrigins[0])
	}
	if cfg.Redis.Enabled {
		t.Fatal("Redis.Enabled = true, want false by default")
	}
	if cfg.Redis.Addr != "127.0.0.1:6379" {
		t.Fatalf("Redis.Addr = %q, want 127.0.0.1:6379", cfg.Redis.Addr)
	}
	if cfg.ASR.Provider != "mock" {
		t.Fatalf("ASR.Provider = %q, want mock", cfg.ASR.Provider)
	}
	if cfg.ASR.TimeoutSeconds != 30 {
		t.Fatalf("ASR.TimeoutSeconds = %d, want 30", cfg.ASR.TimeoutSeconds)
	}
	if !cfg.ASR.UseMock {
		t.Fatal("ASR.UseMock = false, want true by default")
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
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Config.Validate returned error for defaults: %v", err)
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
	t.Setenv("LLM_FALLBACK_TO_MOCK", "false")
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
	if cfg.LLM.FallbackToMock {
		t.Fatal("LLM.FallbackToMock = true, want false")
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

func TestLoadReadsInfrastructureEnvironment(t *testing.T) {
	clearLLMEnv(t)
	clearStorageEnv(t)
	clearInfrastructureEnv(t)
	t.Setenv("REQUEST_TIMEOUT_SECONDS", "12")
	t.Setenv("EXTERNAL_SERVICE_TIMEOUT_SECONDS", "19")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173, https://app.example.com ")
	t.Setenv("CORS_ALLOWED_METHODS", "GET,POST,OPTIONS")
	t.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,X-Request-ID")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	t.Setenv("REDIS_ENABLED", "true")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("REDIS_PASSWORD", "redis-secret")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("REDIS_CONNECT_TIMEOUT_SECONDS", "4")
	t.Setenv("ASR_PROVIDER", "mock")
	t.Setenv("ASR_BASE_URL", "https://asr.example.com/v1")
	t.Setenv("ASR_API_KEY", "asr-test-key")
	t.Setenv("ASR_MODEL", "asr-test-model")
	t.Setenv("ASR_TIMEOUT_SECONDS", "7")
	t.Setenv("ASR_USE_MOCK", "false")

	cfg := Load()

	if cfg.Server.RequestTimeoutSeconds != 12 {
		t.Fatalf("Server.RequestTimeoutSeconds = %d, want 12", cfg.Server.RequestTimeoutSeconds)
	}
	if cfg.ExternalServiceTimeoutSeconds != 19 {
		t.Fatalf("ExternalServiceTimeoutSeconds = %d, want 19", cfg.ExternalServiceTimeoutSeconds)
	}
	if cfg.CORS.AllowedOrigins[1] != "https://app.example.com" {
		t.Fatalf("CORS.AllowedOrigins[1] = %q, want https://app.example.com", cfg.CORS.AllowedOrigins[1])
	}
	if !cfg.CORS.AllowCredentials {
		t.Fatal("CORS.AllowCredentials = false, want true")
	}
	if cfg.Redis.Addr != "redis:6379" {
		t.Fatalf("Redis.Addr = %q, want redis:6379", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "redis-secret" {
		t.Fatalf("Redis.Password = %q, want redis-secret", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 2 {
		t.Fatalf("Redis.DB = %d, want 2", cfg.Redis.DB)
	}
	if cfg.Redis.ConnectTimeoutSeconds != 4 {
		t.Fatalf("Redis.ConnectTimeoutSeconds = %d, want 4", cfg.Redis.ConnectTimeoutSeconds)
	}
	if cfg.ASR.BaseURL != "https://asr.example.com/v1" {
		t.Fatalf("ASR.BaseURL = %q, want https://asr.example.com/v1", cfg.ASR.BaseURL)
	}
	if cfg.ASR.APIKey != "asr-test-key" {
		t.Fatalf("ASR.APIKey = %q, want asr-test-key", cfg.ASR.APIKey)
	}
	if cfg.ASR.Model != "asr-test-model" {
		t.Fatalf("ASR.Model = %q, want asr-test-model", cfg.ASR.Model)
	}
	if cfg.ASR.TimeoutSeconds != 7 {
		t.Fatalf("ASR.TimeoutSeconds = %d, want 7", cfg.ASR.TimeoutSeconds)
	}
	if cfg.ASR.UseMock {
		t.Fatal("ASR.UseMock = true, want false")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Config.Validate returned error for valid infrastructure env: %v", err)
	}
}

func TestExternalServiceTimeoutFeedsLLMAndASRDefaults(t *testing.T) {
	clearLLMEnv(t)
	clearInfrastructureEnv(t)
	t.Setenv("EXTERNAL_SERVICE_TIMEOUT_SECONDS", "11")

	cfg := Load()

	if cfg.LLM.TimeoutSeconds != 11 {
		t.Fatalf("LLM.TimeoutSeconds = %d, want external timeout 11", cfg.LLM.TimeoutSeconds)
	}
	if cfg.ASR.TimeoutSeconds != 11 {
		t.Fatalf("ASR.TimeoutSeconds = %d, want external timeout 11", cfg.ASR.TimeoutSeconds)
	}
}

func TestConfigValidateRejectsEnabledRedisWithoutAddress(t *testing.T) {
	cfg := Config{
		Storage: StorageConfig{Mode: StorageModeMemory},
		Redis:   RedisConfig{Enabled: true},
	}

	err := cfg.Validate()

	if err == nil {
		t.Fatal("Validate returned nil, want missing Redis address error")
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
	t.Setenv("LLM_FALLBACK_TO_MOCK", "not-a-bool")
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
	if !cfg.LLM.FallbackToMock {
		t.Fatal("LLM.FallbackToMock = false, want true for invalid flag")
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
	t.Setenv("LLM_FALLBACK_TO_MOCK", "")
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

func clearInfrastructureEnv(t *testing.T) {
	t.Helper()

	t.Setenv("REQUEST_TIMEOUT_SECONDS", "")
	t.Setenv("EXTERNAL_SERVICE_TIMEOUT_SECONDS", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("CORS_ALLOWED_METHODS", "")
	t.Setenv("CORS_ALLOWED_HEADERS", "")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "")
	t.Setenv("REDIS_ENABLED", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_DB", "")
	t.Setenv("REDIS_CONNECT_TIMEOUT_SECONDS", "")
	t.Setenv("ASR_PROVIDER", "")
	t.Setenv("ASR_BASE_URL", "")
	t.Setenv("ASR_API_KEY", "")
	t.Setenv("ASR_MODEL", "")
	t.Setenv("ASR_TIMEOUT_SECONDS", "")
	t.Setenv("ASR_USE_MOCK", "")
}
