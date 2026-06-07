package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

const (
	// StorageModeMemory 表示使用内存仓库，适合本地开发和自动测试。
	StorageModeMemory = "memory"
	// StorageModeMySQL 表示使用 MySQL 仓库存储核心训练数据。
	StorageModeMySQL = "mysql"
)

var (
	// ErrStorageModeUnsupported 表示配置了不支持的存储模式。
	ErrStorageModeUnsupported = errors.New("unsupported storage mode")
	// ErrMySQLDSNRequired 表示 MySQL 模式缺少 DSN。
	ErrMySQLDSNRequired = errors.New("mysql dsn required")
	// ErrRedisAddrRequired 表示启用 Redis 预留连接时缺少地址。
	ErrRedisAddrRequired = errors.New("redis addr required")
)

type Config struct {
	Port                          string
	Server                        ServerConfig
	CORS                          CORSConfig
	LLM                           LLMConfig
	ASR                           ASRConfig
	Feedback                      FeedbackConfig
	Storage                       StorageConfig
	Redis                         RedisConfig
	ExternalServiceTimeoutSeconds int
}

type ServerConfig struct {
	RequestTimeoutSeconds  int
	RequestBodyLimitBytes  int
	RateLimitRequests      int
	RateLimitWindowSeconds int
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type LLMConfig struct {
	Provider       string
	BaseURL        string
	APIKey         string
	Model          string
	TimeoutSeconds int
	UseMock        bool
	FallbackToMock bool
}

type ASRConfig struct {
	Provider               string
	BaseURL                string
	APIKey                 string
	Model                  string
	TimeoutSeconds         int
	UseMock                bool
	TencentAppID           string
	TencentSecretID        string
	TencentSecretKey       string
	TencentEngineType      string
	TencentVoiceFormat     string
	TencentHotwordID       string
	TencentHotwordList     string
	TencentCustomizationID string
	TencentFilterDirty     int
	TencentFilterModal     int
	TencentFilterPunc      int
	TencentConvertNumMode  int
	TencentWordInfo        int
}

type FeedbackConfig struct {
	CorrectionUseMock bool
	ScoringUseMock    bool
	SummaryUseMock    bool
	FailOpen          bool
}

type StorageConfig struct {
	Mode     string
	MySQLDSN string
}

type RedisConfig struct {
	Enabled               bool
	Addr                  string
	Password              string
	DB                    int
	ConnectTimeoutSeconds int
}

func Load() Config {
	loadDotEnvIntoProcessEnv(".env")

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	externalServiceTimeoutSeconds := positiveIntEnv("EXTERNAL_SERVICE_TIMEOUT_SECONDS", 30)
	llmUseMock := boolEnv("LLM_USE_MOCK", true)
	llmFallbackToMock := boolEnv("LLM_FALLBACK_TO_MOCK", llmUseMock)

	return Config{
		Port: port,
		Server: ServerConfig{
			RequestTimeoutSeconds:  positiveIntEnv("REQUEST_TIMEOUT_SECONDS", 30),
			RequestBodyLimitBytes:  positiveIntEnv("REQUEST_BODY_LIMIT_BYTES", 12*1024*1024),
			RateLimitRequests:      positiveIntEnv("RATE_LIMIT_REQUESTS", 120),
			RateLimitWindowSeconds: positiveIntEnv("RATE_LIMIT_WINDOW_SECONDS", 60),
		},
		CORS: CORSConfig{
			AllowedOrigins:   listEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
			AllowedMethods:   listEnv("CORS_ALLOWED_METHODS", []string{"GET", "POST", "OPTIONS"}),
			AllowedHeaders:   listEnv("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
			AllowCredentials: boolEnv("CORS_ALLOW_CREDENTIALS", false),
		},
		LLM: LLMConfig{
			Provider:       stringEnv("LLM_PROVIDER", "openai-compatible"),
			BaseURL:        strings.TrimSpace(os.Getenv("LLM_BASE_URL")),
			APIKey:         strings.TrimSpace(os.Getenv("LLM_API_KEY")),
			Model:          strings.TrimSpace(os.Getenv("LLM_MODEL")),
			TimeoutSeconds: positiveIntEnv("LLM_TIMEOUT_SECONDS", externalServiceTimeoutSeconds),
			UseMock:        llmUseMock,
			FallbackToMock: llmFallbackToMock,
		},
		ASR: ASRConfig{
			Provider:               stringEnv("ASR_PROVIDER", "mock"),
			BaseURL:                strings.TrimSpace(os.Getenv("ASR_BASE_URL")),
			APIKey:                 strings.TrimSpace(os.Getenv("ASR_API_KEY")),
			Model:                  strings.TrimSpace(os.Getenv("ASR_MODEL")),
			TimeoutSeconds:         positiveIntEnv("ASR_TIMEOUT_SECONDS", externalServiceTimeoutSeconds),
			UseMock:                boolEnv("ASR_USE_MOCK", true),
			TencentAppID:           strings.TrimSpace(os.Getenv("TENCENT_ASR_APP_ID")),
			TencentSecretID:        strings.TrimSpace(os.Getenv("TENCENT_ASR_SECRET_ID")),
			TencentSecretKey:       strings.TrimSpace(os.Getenv("TENCENT_ASR_SECRET_KEY")),
			TencentEngineType:      strings.TrimSpace(os.Getenv("TENCENT_ASR_ENGINE_TYPE")),
			TencentVoiceFormat:     stringEnv("TENCENT_ASR_VOICE_FORMAT", "ogg-opus"),
			TencentHotwordID:       strings.TrimSpace(os.Getenv("TENCENT_ASR_HOTWORD_ID")),
			TencentHotwordList:     strings.TrimSpace(os.Getenv("TENCENT_ASR_HOTWORD_LIST")),
			TencentCustomizationID: strings.TrimSpace(os.Getenv("TENCENT_ASR_CUSTOMIZATION_ID")),
			TencentFilterDirty:     nonNegativeIntEnv("TENCENT_ASR_FILTER_DIRTY", 0),
			TencentFilterModal:     nonNegativeIntEnv("TENCENT_ASR_FILTER_MODAL", 0),
			TencentFilterPunc:      nonNegativeIntEnv("TENCENT_ASR_FILTER_PUNC", 0),
			TencentConvertNumMode:  nonNegativeIntEnv("TENCENT_ASR_CONVERT_NUM_MODE", 1),
			TencentWordInfo:        nonNegativeIntEnv("TENCENT_ASR_WORD_INFO", 0),
		},
		Feedback: FeedbackConfig{
			CorrectionUseMock: boolEnv("CORRECTION_USE_MOCK", true),
			ScoringUseMock:    boolEnv("SCORING_USE_MOCK", true),
			SummaryUseMock:    boolEnv("SUMMARY_USE_MOCK", true),
			FailOpen:          boolEnv("FEEDBACK_FAIL_OPEN", true),
		},
		Storage: StorageConfig{
			Mode:     normalizeStorageMode(stringEnv("STORAGE_MODE", StorageModeMemory)),
			MySQLDSN: strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		},
		Redis: RedisConfig{
			Enabled:               boolEnv("REDIS_ENABLED", false),
			Addr:                  strings.TrimSpace(stringEnv("REDIS_ADDR", "127.0.0.1:6379")),
			Password:              strings.TrimSpace(os.Getenv("REDIS_PASSWORD")),
			DB:                    nonNegativeIntEnv("REDIS_DB", 0),
			ConnectTimeoutSeconds: positiveIntEnv("REDIS_CONNECT_TIMEOUT_SECONDS", externalServiceTimeoutSeconds),
		},
		ExternalServiceTimeoutSeconds: externalServiceTimeoutSeconds,
	}
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func (c Config) Validate() error {
	if err := c.Storage.Validate(); err != nil {
		return err
	}
	if err := c.Redis.Validate(); err != nil {
		return err
	}

	return nil
}

func (c LLMConfig) HasRequiredFields() bool {
	return strings.TrimSpace(c.BaseURL) != "" &&
		strings.TrimSpace(c.APIKey) != "" &&
		strings.TrimSpace(c.Model) != ""
}

func (c ASRConfig) HasRequiredFields() bool {
	return strings.TrimSpace(c.BaseURL) != "" &&
		strings.TrimSpace(c.APIKey) != "" &&
		strings.TrimSpace(c.Model) != ""
}

func (c ASRConfig) HasTencentRequiredFields() bool {
	return strings.TrimSpace(c.TencentAppID) != "" &&
		strings.TrimSpace(c.TencentSecretID) != "" &&
		strings.TrimSpace(c.TencentSecretKey) != "" &&
		strings.TrimSpace(c.TencentEngineType) != ""
}

func (c StorageConfig) Validate() error {
	mode := normalizeStorageMode(c.Mode)
	if mode == "" {
		mode = StorageModeMemory
	}

	switch mode {
	case StorageModeMemory:
		return nil
	case StorageModeMySQL:
		if strings.TrimSpace(c.MySQLDSN) == "" {
			return ErrMySQLDSNRequired
		}

		return nil
	default:
		return ErrStorageModeUnsupported
	}
}

func (c StorageConfig) IsMySQL() bool {
	return normalizeStorageMode(c.Mode) == StorageModeMySQL
}

func (c RedisConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.Addr) == "" {
		return ErrRedisAddrRequired
	}

	return nil
}

func normalizeStorageMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
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

func nonNegativeIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
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

func listEnv(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return append([]string(nil), fallback...)
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return append([]string(nil), fallback...)
	}

	return items
}

func loadDotEnvIntoProcessEnv(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, rawLine := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" || strings.TrimSpace(os.Getenv(key)) != "" {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
	}
}
