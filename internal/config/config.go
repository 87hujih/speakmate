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
)

type Config struct {
	Port     string
	LLM      LLMConfig
	Feedback FeedbackConfig
	Storage  StorageConfig
}

type LLMConfig struct {
	Provider       string
	BaseURL        string
	APIKey         string
	Model          string
	TimeoutSeconds int
	UseMock        bool
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
