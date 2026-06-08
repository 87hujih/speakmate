package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsToPort8080(t *testing.T) {
	t.Setenv("APP_PORT", "")
	clearLLMEnv(t)
	clearStorageEnv(t)
	withoutDotEnv(t)

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
	if cfg.Server.RequestBodyLimitBytes != 12*1024*1024 {
		t.Fatalf("Server.RequestBodyLimitBytes = %d, want 12MiB", cfg.Server.RequestBodyLimitBytes)
	}
	if cfg.Server.RateLimitRequests != 120 {
		t.Fatalf("Server.RateLimitRequests = %d, want 120", cfg.Server.RateLimitRequests)
	}
	if cfg.Server.RateLimitWindowSeconds != 60 {
		t.Fatalf("Server.RateLimitWindowSeconds = %d, want 60", cfg.Server.RateLimitWindowSeconds)
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
	withoutDotEnv(t)
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
	withoutDotEnv(t)
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
	withoutDotEnv(t)
	t.Setenv("REQUEST_TIMEOUT_SECONDS", "12")
	t.Setenv("REQUEST_BODY_LIMIT_BYTES", "2048")
	t.Setenv("RATE_LIMIT_REQUESTS", "9")
	t.Setenv("RATE_LIMIT_WINDOW_SECONDS", "3")
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
	t.Setenv("ASR_MOCK_TRANSCRIPT", "I built a billing dashboard.")
	t.Setenv("TENCENT_ASR_APP_ID", "1250000000")
	t.Setenv("TENCENT_ASR_SECRET_ID", "test-secret-id")
	t.Setenv("TENCENT_ASR_SECRET_KEY", "test-secret-key")
	t.Setenv("TENCENT_ASR_ENGINE_TYPE", "16k_en")
	t.Setenv("TENCENT_ASR_VOICE_FORMAT", "ogg-opus")
	t.Setenv("TENCENT_ASR_HOTWORD_ID", "hotword-id")
	t.Setenv("TENCENT_ASR_HOTWORD_LIST", "cloud word")
	t.Setenv("TENCENT_ASR_CUSTOMIZATION_ID", "custom-id")
	t.Setenv("TENCENT_ASR_FILTER_DIRTY", "1")
	t.Setenv("TENCENT_ASR_FILTER_MODAL", "1")
	t.Setenv("TENCENT_ASR_FILTER_PUNC", "1")
	t.Setenv("TENCENT_ASR_CONVERT_NUM_MODE", "0")
	t.Setenv("TENCENT_ASR_WORD_INFO", "2")

	cfg := Load()

	if cfg.Server.RequestTimeoutSeconds != 12 {
		t.Fatalf("Server.RequestTimeoutSeconds = %d, want 12", cfg.Server.RequestTimeoutSeconds)
	}
	if cfg.Server.RequestBodyLimitBytes != 2048 {
		t.Fatalf("Server.RequestBodyLimitBytes = %d, want 2048", cfg.Server.RequestBodyLimitBytes)
	}
	if cfg.Server.RateLimitRequests != 9 {
		t.Fatalf("Server.RateLimitRequests = %d, want 9", cfg.Server.RateLimitRequests)
	}
	if cfg.Server.RateLimitWindowSeconds != 3 {
		t.Fatalf("Server.RateLimitWindowSeconds = %d, want 3", cfg.Server.RateLimitWindowSeconds)
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
	if cfg.ASR.MockTranscript != "I built a billing dashboard." {
		t.Fatalf("ASR.MockTranscript = %q, want configured mock transcript", cfg.ASR.MockTranscript)
	}
	if cfg.ASR.TencentAppID != "1250000000" {
		t.Fatalf("ASR.TencentAppID = %q, want configured app id", cfg.ASR.TencentAppID)
	}
	if cfg.ASR.TencentSecretID != "test-secret-id" {
		t.Fatalf("ASR.TencentSecretID = %q, want configured secret id", cfg.ASR.TencentSecretID)
	}
	if cfg.ASR.TencentSecretKey != "test-secret-key" {
		t.Fatalf("ASR.TencentSecretKey = %q, want configured secret key", cfg.ASR.TencentSecretKey)
	}
	if cfg.ASR.TencentEngineType != "16k_en" {
		t.Fatalf("ASR.TencentEngineType = %q, want 16k_en", cfg.ASR.TencentEngineType)
	}
	if cfg.ASR.TencentVoiceFormat != "ogg-opus" {
		t.Fatalf("ASR.TencentVoiceFormat = %q, want ogg-opus", cfg.ASR.TencentVoiceFormat)
	}
	if cfg.ASR.TencentHotwordID != "hotword-id" {
		t.Fatalf("ASR.TencentHotwordID = %q, want hotword-id", cfg.ASR.TencentHotwordID)
	}
	if cfg.ASR.TencentHotwordList != "cloud word" {
		t.Fatalf("ASR.TencentHotwordList = %q, want cloud word", cfg.ASR.TencentHotwordList)
	}
	if cfg.ASR.TencentCustomizationID != "custom-id" {
		t.Fatalf("ASR.TencentCustomizationID = %q, want custom-id", cfg.ASR.TencentCustomizationID)
	}
	if cfg.ASR.TencentFilterDirty != 1 {
		t.Fatalf("ASR.TencentFilterDirty = %d, want 1", cfg.ASR.TencentFilterDirty)
	}
	if cfg.ASR.TencentFilterModal != 1 {
		t.Fatalf("ASR.TencentFilterModal = %d, want 1", cfg.ASR.TencentFilterModal)
	}
	if cfg.ASR.TencentFilterPunc != 1 {
		t.Fatalf("ASR.TencentFilterPunc = %d, want 1", cfg.ASR.TencentFilterPunc)
	}
	if cfg.ASR.TencentConvertNumMode != 0 {
		t.Fatalf("ASR.TencentConvertNumMode = %d, want 0", cfg.ASR.TencentConvertNumMode)
	}
	if cfg.ASR.TencentWordInfo != 2 {
		t.Fatalf("ASR.TencentWordInfo = %d, want 2", cfg.ASR.TencentWordInfo)
	}
	if !cfg.ASR.HasTencentRequiredFields() {
		t.Fatal("ASR.HasTencentRequiredFields() = false, want true")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Config.Validate returned error for valid infrastructure env: %v", err)
	}
}

func TestLoadReadsDotEnvWhenProcessEnvironmentUnset(t *testing.T) {
	clearLLMEnv(t)
	clearStorageEnv(t)
	clearInfrastructureEnv(t)
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte(`
ASR_PROVIDER=tencent
ASR_USE_MOCK=false
TENCENT_ASR_APP_ID=1250000000
TENCENT_ASR_SECRET_ID=test-secret-id
TENCENT_ASR_SECRET_KEY=test-secret-key
TENCENT_ASR_ENGINE_TYPE=16k_en
`), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg := Load()

	if cfg.ASR.Provider != "tencent" {
		t.Fatalf("ASR.Provider = %q, want tencent", cfg.ASR.Provider)
	}
	if cfg.ASR.UseMock {
		t.Fatal("ASR.UseMock = true, want false from .env")
	}
	if !cfg.ASR.HasTencentRequiredFields() {
		t.Fatal("ASR.HasTencentRequiredFields() = false, want true from .env")
	}
}

func TestLoadPrefersProcessEnvironmentOverDotEnv(t *testing.T) {
	clearLLMEnv(t)
	clearStorageEnv(t)
	clearInfrastructureEnv(t)
	t.Setenv("ASR_PROVIDER", "mock")
	t.Setenv("ASR_USE_MOCK", "true")
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte(`
ASR_PROVIDER=tencent
ASR_USE_MOCK=false
`), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg := Load()

	if cfg.ASR.Provider != "mock" {
		t.Fatalf("ASR.Provider = %q, want process env mock", cfg.ASR.Provider)
	}
	if !cfg.ASR.UseMock {
		t.Fatal("ASR.UseMock = false, want process env true")
	}
}

func TestLoadFindsDotEnvInParentDirectory(t *testing.T) {
	clearLLMEnv(t)
	clearStorageEnv(t)
	clearInfrastructureEnv(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(`
ASR_PROVIDER=tencent
ASR_USE_MOCK=false
TENCENT_ASR_APP_ID=1250000000
TENCENT_ASR_SECRET_ID=test-secret-id
TENCENT_ASR_SECRET_KEY=test-secret-key
TENCENT_ASR_ENGINE_TYPE=16k_en
`), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	binDir := filepath.Join(root, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatalf("make bin dir: %v", err)
	}
	t.Chdir(binDir)

	cfg := Load()

	if cfg.ASR.Provider != "tencent" {
		t.Fatalf("ASR.Provider = %q, want tencent from parent .env", cfg.ASR.Provider)
	}
	if cfg.ASR.UseMock {
		t.Fatal("ASR.UseMock = true, want false from parent .env")
	}
}

func TestLoadReadsTencentEngineModelTypeAlias(t *testing.T) {
	clearLLMEnv(t)
	clearStorageEnv(t)
	clearInfrastructureEnv(t)
	withoutDotEnv(t)
	t.Setenv("ASR_PROVIDER", "tencent")
	t.Setenv("ASR_USE_MOCK", "false")
	t.Setenv("TENCENT_ASR_APP_ID", "1250000000")
	t.Setenv("TENCENT_ASR_SECRET_ID", "test-secret-id")
	t.Setenv("TENCENT_ASR_SECRET_KEY", "test-secret-key")
	t.Setenv("TENCENT_ASR_ENGINE_MODEL_TYPE", "16k_zh")

	cfg := Load()

	if cfg.ASR.TencentEngineType != "16k_zh" {
		t.Fatalf("ASR.TencentEngineType = %q, want alias engine model type", cfg.ASR.TencentEngineType)
	}
	if !cfg.ASR.HasTencentRequiredFields() {
		t.Fatal("ASR.HasTencentRequiredFields() = false, want true from engine alias")
	}
}

func TestLoadDisablesLLMFallbackByDefaultWhenRealLLMIsConfigured(t *testing.T) {
	clearLLMEnv(t)
	clearStorageEnv(t)
	clearInfrastructureEnv(t)
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte(`
LLM_PROVIDER=openai-compatible
LLM_BASE_URL=https://llm.example.com/v1
LLM_API_KEY=test-key
LLM_MODEL=test-model
LLM_USE_MOCK=false
CORRECTION_USE_MOCK=false
SCORING_USE_MOCK=false
SUMMARY_USE_MOCK=false
`), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg := Load()

	if cfg.LLM.UseMock {
		t.Fatal("LLM.UseMock = true, want false from .env")
	}
	if cfg.LLM.FallbackToMock {
		t.Fatal("LLM.FallbackToMock = true, want false when real LLM mode is requested")
	}
	if cfg.Feedback.CorrectionUseMock {
		t.Fatal("Feedback.CorrectionUseMock = true, want false from .env")
	}
	if cfg.Feedback.ScoringUseMock {
		t.Fatal("Feedback.ScoringUseMock = true, want false from .env")
	}
	if !cfg.LLM.HasRequiredFields() {
		t.Fatal("LLM.HasRequiredFields() = false, want true from .env")
	}
}

func TestASRConfigHasTencentRequiredFields(t *testing.T) {
	valid := ASRConfig{
		TencentAppID:      "1250000000",
		TencentSecretID:   "secret-id",
		TencentSecretKey:  "secret-key",
		TencentEngineType: "16k_en",
	}
	if !valid.HasTencentRequiredFields() {
		t.Fatal("HasTencentRequiredFields() = false, want true for complete Tencent config")
	}

	tests := []struct {
		name string
		cfg  ASRConfig
	}{
		{name: "missing app id", cfg: ASRConfig{TencentSecretID: "secret-id", TencentSecretKey: "secret-key", TencentEngineType: "16k_en"}},
		{name: "missing secret id", cfg: ASRConfig{TencentAppID: "1250000000", TencentSecretKey: "secret-key", TencentEngineType: "16k_en"}},
		{name: "missing secret key", cfg: ASRConfig{TencentAppID: "1250000000", TencentSecretID: "secret-id", TencentEngineType: "16k_en"}},
		{name: "missing engine type", cfg: ASRConfig{TencentAppID: "1250000000", TencentSecretID: "secret-id", TencentSecretKey: "secret-key"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg.HasTencentRequiredFields() {
				t.Fatal("HasTencentRequiredFields() = true, want false")
			}
		})
	}
}

func TestExternalServiceTimeoutFeedsLLMAndASRDefaults(t *testing.T) {
	clearLLMEnv(t)
	clearInfrastructureEnv(t)
	withoutDotEnv(t)
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
	withoutDotEnv(t)
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

func withoutDotEnv(t *testing.T) {
	t.Helper()

	t.Chdir(t.TempDir())
}

func clearStorageEnv(t *testing.T) {
	t.Helper()

	t.Setenv("STORAGE_MODE", "")
	t.Setenv("MYSQL_DSN", "")
}

func clearInfrastructureEnv(t *testing.T) {
	t.Helper()

	t.Setenv("REQUEST_TIMEOUT_SECONDS", "")
	t.Setenv("REQUEST_BODY_LIMIT_BYTES", "")
	t.Setenv("RATE_LIMIT_REQUESTS", "")
	t.Setenv("RATE_LIMIT_WINDOW_SECONDS", "")
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
	t.Setenv("ASR_MOCK_TRANSCRIPT", "")
	t.Setenv("TENCENT_ASR_APP_ID", "")
	t.Setenv("TENCENT_ASR_SECRET_ID", "")
	t.Setenv("TENCENT_ASR_SECRET_KEY", "")
	t.Setenv("TENCENT_ASR_ENGINE_TYPE", "")
	t.Setenv("TENCENT_ASR_ENGINE_MODEL_TYPE", "")
	t.Setenv("TENCENT_ASR_VOICE_FORMAT", "")
	t.Setenv("TENCENT_ASR_HOTWORD_ID", "")
	t.Setenv("TENCENT_ASR_HOTWORD_LIST", "")
	t.Setenv("TENCENT_ASR_CUSTOMIZATION_ID", "")
	t.Setenv("TENCENT_ASR_FILTER_DIRTY", "")
	t.Setenv("TENCENT_ASR_FILTER_MODAL", "")
	t.Setenv("TENCENT_ASR_FILTER_PUNC", "")
	t.Setenv("TENCENT_ASR_CONVERT_NUM_MODE", "")
	t.Setenv("TENCENT_ASR_WORD_INFO", "")
}
