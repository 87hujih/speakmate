package router

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/agent"
	"speakmate/internal/config"
	"speakmate/internal/handler"
	infraasr "speakmate/internal/infra/asr"
	"speakmate/internal/infra/database"
	"speakmate/internal/infra/llm"
	infraredis "speakmate/internal/infra/redis"
	"speakmate/internal/middleware"
	"speakmate/internal/repository"
	"speakmate/internal/service"
	"speakmate/internal/state"
	"speakmate/internal/stream"
)

// New 创建并配置 Gin 路由引擎。
func New(configs ...config.Config) *gin.Engine {
	engine, err := NewWithError(configs...)
	if err != nil {
		panic(err)
	}

	return engine
}

// NewWithError 创建并配置 Gin 路由引擎，初始化失败时返回明确错误。
func NewWithError(configs ...config.Config) (*gin.Engine, error) {
	cfg := config.Load()
	if len(configs) > 0 {
		cfg = configs[0]
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	logger := log.New(os.Stdout, "", log.LstdFlags)
	engine.Use(
		middleware.Recover(logger),
		middleware.RequestLogger(logger),
		middleware.CORS(cfg.CORS),
		middleware.RateLimit(cfg.Server.RateLimitRequests, time.Duration(cfg.Server.RateLimitWindowSeconds)*time.Second),
		middleware.BodySizeLimit(cfg.Server.RequestBodyLimitBytes),
		middleware.RequestTimeout(time.Duration(cfg.Server.RequestTimeoutSeconds)*time.Second),
	)
	engine.GET("/health", handler.Health)

	scenarioRepo, sessionRepo, feedbackRepo, reportRepo, err := newRepositories(cfg)
	if err != nil {
		return nil, err
	}
	stateStore, eventBus, err := newRealtimeState(cfg)
	if err != nil {
		return nil, err
	}
	scenarioService := service.NewScenarioService(scenarioRepo)
	scenarioHandler := handler.NewScenarioHandler(scenarioService)
	sessionService := service.NewSessionService(
		scenarioService,
		sessionRepo,
		service.WithConversationAgent(NewConversationAgent(cfg)),
		service.WithFeedbackRepository(feedbackRepo),
		service.WithCorrectionAgent(NewCorrectionAgent(cfg)),
		service.WithScoringAgent(NewScoringAgent(cfg)),
		service.WithFeedbackFailOpen(cfg.Feedback.FailOpen),
		service.WithEventPublisher(eventBus),
		service.WithStateStore(stateStore),
	)
	sessionHandler := handler.NewSessionHandler(sessionService)
	messageHandler := handler.NewMessageHandler(sessionService)
	asrClient, err := NewASRClient(cfg)
	if err != nil {
		return nil, err
	}
	audioService := service.NewAudioService(sessionService, asrClient)
	audioHandler := handler.NewAudioHandler(audioService)
	audioStreamService := service.NewAudioStreamService(
		sessionService,
		asrClient,
		service.WithAudioStreamPartialTranscription(audioStreamPartialTranscriptionEnabled(cfg)),
		service.WithAudioStreamStateStore(stateStore),
	)
	audioWebSocketHandler := handler.NewAudioWebSocketHandler(audioStreamService, cfg.CORS)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	feedbackHandler := handler.NewFeedbackHandler(feedbackService)
	reportService := service.NewReportService(
		scenarioService,
		sessionRepo,
		feedbackRepo,
		reportRepo,
		service.WithSummaryAgent(NewSummaryAgent(cfg)),
		service.WithReportEventPublisher(eventBus),
	)
	reportHandler := handler.NewReportHandler(reportService)
	historyService := service.NewHistoryService(scenarioService, sessionRepo, feedbackRepo, reportRepo)
	historyHandler := handler.NewHistoryHandler(historyService)
	historyInsightsService := service.NewHistoryInsightsService(scenarioService, sessionRepo, feedbackRepo, reportRepo)
	historyInsightsHandler := handler.NewHistoryInsightsHandler(historyInsightsService)
	streamHandler := handler.NewStreamHandler(eventBus)

	// v1 API 路由组承载场景、训练 Session 和消息等后续接口。
	api := engine.Group("/api/v1")
	api.GET("/scenarios", scenarioHandler.List)
	api.GET("/scenarios/:id", scenarioHandler.Detail)
	api.GET("/sessions", historyHandler.List)
	api.GET("/users/:user_id/sessions", historyHandler.ListByUser)
	api.GET("/history/insights", historyInsightsHandler.Get)
	api.POST("/sessions", sessionHandler.Create)
	api.GET("/sessions/:id", sessionHandler.Detail)
	api.GET("/sessions/:id/stream", streamHandler.Stream)
	api.POST("/sessions/:id/finish", sessionHandler.Finish)
	api.POST("/sessions/:id/messages", messageHandler.Send)
	api.POST("/sessions/:id/audio", audioHandler.Upload)
	api.GET("/sessions/:id/audio/ws", audioWebSocketHandler.Stream)
	api.GET("/sessions/:id/corrections", feedbackHandler.ListSessionCorrections)
	api.GET("/sessions/:id/scores", feedbackHandler.GetSessionScore)
	api.POST("/sessions/:id/report", reportHandler.Generate)
	api.GET("/sessions/:id/report", reportHandler.Get)
	api.GET("/messages/:message_id/corrections", feedbackHandler.GetMessageCorrection)

	return engine, nil
}

// realtimeEventBus 聚合事件订阅和发布能力。
type realtimeEventBus interface {
	service.EventPublisher
	handler.EventSubscriber
}

// newRealtimeState 创建训练实时状态存储。
func newRealtimeState(cfg config.Config) (state.SessionStateStore, realtimeEventBus, error) {
	if cfg.Redis.Enabled {
		client, err := infraredis.OpenClient(context.Background(), cfg.Redis)
		if err != nil {
			return nil, nil, err
		}

		return state.NewRedisSessionStateStore(client), stream.NewRedisBus(client), nil
	}

	return state.NewMemorySessionStateStore(), stream.NewBus(), nil
}

// sessionStore 聚合 Session 仓库和历史查询能力。
type sessionStore interface {
	service.SessionRepository
	service.ReportSessionReader
	service.HistorySessionRepository
	service.HistoryInsightSessionRepository
}

// feedbackStore 聚合反馈仓库读写能力。
type feedbackStore interface {
	service.FeedbackRepository
	service.ReportFeedbackReader
}

// reportStore 聚合报告仓库读写能力。
type reportStore interface {
	service.ReportRepository
	service.HistoryReportRepository
}

// newRepositories 根据配置创建长期数据仓库。
func newRepositories(cfg config.Config) (service.ScenarioRepository, sessionStore, feedbackStore, reportStore, error) {
	if cfg.Storage.IsMySQL() {
		db, err := database.OpenMySQL(context.Background(), cfg.Storage)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		return repository.NewMySQLScenarioRepository(db),
			repository.NewMySQLSessionRepository(db),
			repository.NewMySQLFeedbackRepository(db),
			repository.NewMySQLReportRepository(db),
			nil
	}

	return repository.NewMemoryScenarioRepository(),
		repository.NewMemorySessionRepository(),
		repository.NewMemoryFeedbackRepository(),
		repository.NewMemoryReportRepository(),
		nil
}

// NewConversationAgent 根据配置创建对话 Agent。
func NewConversationAgent(cfg config.Config) agent.ConversationAgent {
	if cfg.LLM.UseMock {
		return agent.NewUnavailableConversationAgent("真实 LLM 对话已禁用，因为 LLM_USE_MOCK 为 true")
	}
	if !cfg.LLM.HasRequiredFields() {
		return agent.NewUnavailableConversationAgent("LLM BaseURL、API Key 和模型不能为空")
	}
	if !strings.EqualFold(cfg.LLM.Provider, "openai-compatible") {
		return agent.NewUnavailableConversationAgent("不支持的对话 LLM Provider：" + cfg.LLM.Provider)
	}

	client, err := llm.NewOpenAICompatibleClient(cfg.LLM)
	if err != nil {
		return agent.NewUnavailableConversationAgent(err.Error())
	}

	return agent.NewLLMConversationAgent(client)
}

// NewCorrectionAgent 根据配置创建纠错 Agent。
func NewCorrectionAgent(cfg config.Config) agent.CorrectionAgent {
	if cfg.LLM.UseMock || cfg.Feedback.CorrectionUseMock || !cfg.LLM.HasRequiredFields() {
		return agent.NewMockCorrectionAgent()
	}
	if !strings.EqualFold(cfg.LLM.Provider, "openai-compatible") {
		return agent.NewMockCorrectionAgent()
	}

	client, err := llm.NewOpenAICompatibleClient(cfg.LLM)
	if err != nil {
		return agent.NewMockCorrectionAgent()
	}

	opts := []agent.LLMCorrectionOption{}
	if cfg.LLM.FallbackToMock {
		opts = append(opts, agent.WithCorrectionFallbackAgent(agent.NewMockCorrectionAgent()))
	}

	return agent.NewLLMCorrectionAgent(client, opts...)
}

// NewScoringAgent 根据配置创建评分 Agent。
func NewScoringAgent(cfg config.Config) agent.ScoringAgent {
	if cfg.LLM.UseMock || cfg.Feedback.ScoringUseMock || !cfg.LLM.HasRequiredFields() {
		return agent.NewMockScoringAgent()
	}
	if !strings.EqualFold(cfg.LLM.Provider, "openai-compatible") {
		return agent.NewMockScoringAgent()
	}

	client, err := llm.NewOpenAICompatibleClient(cfg.LLM)
	if err != nil {
		return agent.NewMockScoringAgent()
	}

	opts := []agent.LLMScoringOption{}
	if cfg.LLM.FallbackToMock {
		opts = append(opts, agent.WithScoringFallbackAgent(agent.NewMockScoringAgent()))
	}

	return agent.NewLLMScoringAgent(client, opts...)
}

// NewSummaryAgent 根据配置创建报告摘要 Agent。
func NewSummaryAgent(cfg config.Config) agent.SummaryAgent {
	if cfg.LLM.UseMock || cfg.Feedback.SummaryUseMock || !cfg.LLM.HasRequiredFields() {
		return agent.NewMockSummaryAgent()
	}
	if !strings.EqualFold(cfg.LLM.Provider, "openai-compatible") {
		return agent.NewMockSummaryAgent()
	}

	client, err := llm.NewOpenAICompatibleClient(cfg.LLM)
	if err != nil {
		return agent.NewMockSummaryAgent()
	}

	opts := []agent.LLMSummaryOption{}
	if cfg.LLM.FallbackToMock {
		opts = append(opts, agent.WithSummaryFallbackAgent(agent.NewMockSummaryAgent()))
	}

	return agent.NewLLMSummaryAgent(client, opts...)
}

// NewASRClient 根据配置创建 ASR 客户端。
func NewASRClient(cfg config.Config) (agent.ASRClient, error) {
	if cfg.ASR.UseMock || strings.EqualFold(cfg.ASR.Provider, "mock") {
		return agent.NewMockASRClient(agent.WithMockASRTranscript(cfg.ASR.MockTranscript)), nil
	}
	if strings.EqualFold(cfg.ASR.Provider, "tencent") {
		return infraasr.NewTencentFlashClient(cfg.ASR)
	}

	return nil, fmt.Errorf("不支持的 ASR Provider：%s", cfg.ASR.Provider)
}

// audioStreamPartialTranscriptionEnabled 判断实时音频是否启用 partial 转写。
func audioStreamPartialTranscriptionEnabled(cfg config.Config) bool {
	return cfg.ASR.UseMock || strings.EqualFold(cfg.ASR.Provider, "mock")
}
