package router

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"speakmate/internal/agent"
	"speakmate/internal/config"
	"speakmate/internal/handler"
	"speakmate/internal/infra/database"
	"speakmate/internal/infra/llm"
	"speakmate/internal/repository"
	"speakmate/internal/service"
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
	if err := cfg.Storage.Validate(); err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/health", handler.Health)

	scenarioRepo, sessionRepo, feedbackRepo, reportRepo, err := newRepositories(cfg)
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
	)
	sessionHandler := handler.NewSessionHandler(sessionService)
	messageHandler := handler.NewMessageHandler(sessionService)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	feedbackHandler := handler.NewFeedbackHandler(feedbackService)
	reportService := service.NewReportService(
		scenarioService,
		sessionRepo,
		feedbackRepo,
		reportRepo,
		service.WithSummaryAgent(NewSummaryAgent(cfg)),
	)
	reportHandler := handler.NewReportHandler(reportService)
	historyService := service.NewHistoryService(scenarioService, sessionRepo, feedbackRepo, reportRepo)
	historyHandler := handler.NewHistoryHandler(historyService)

	// v1 API 路由组承载场景、训练 Session 和消息等后续接口。
	api := engine.Group("/api/v1")
	api.GET("/scenarios", scenarioHandler.List)
	api.GET("/scenarios/:id", scenarioHandler.Detail)
	api.GET("/sessions", historyHandler.List)
	api.GET("/users/:user_id/sessions", historyHandler.ListByUser)
	api.POST("/sessions", sessionHandler.Create)
	api.GET("/sessions/:id", sessionHandler.Detail)
	api.POST("/sessions/:id/finish", sessionHandler.Finish)
	api.POST("/sessions/:id/messages", messageHandler.Send)
	api.GET("/sessions/:id/corrections", feedbackHandler.ListSessionCorrections)
	api.GET("/sessions/:id/scores", feedbackHandler.GetSessionScore)
	api.POST("/sessions/:id/report", reportHandler.Generate)
	api.GET("/sessions/:id/report", reportHandler.Get)
	api.GET("/messages/:message_id/corrections", feedbackHandler.GetMessageCorrection)

	return engine, nil
}

type sessionStore interface {
	service.SessionRepository
	service.ReportSessionReader
	service.HistorySessionRepository
}

type feedbackStore interface {
	service.FeedbackRepository
	service.ReportFeedbackReader
}

type reportStore interface {
	service.ReportRepository
	service.HistoryReportRepository
}

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

func NewConversationAgent(cfg config.Config) agent.ConversationAgent {
	if cfg.LLM.UseMock || !cfg.LLM.HasRequiredFields() {
		return agent.NewMockConversationAgent()
	}
	if !strings.EqualFold(cfg.LLM.Provider, "openai-compatible") {
		return agent.NewMockConversationAgent()
	}

	client, err := llm.NewOpenAICompatibleClient(cfg.LLM)
	if err != nil {
		return agent.NewMockConversationAgent()
	}

	return agent.NewLLMConversationAgent(client, agent.WithFallbackAgent(agent.NewMockConversationAgent()))
}

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

	return agent.NewLLMCorrectionAgent(client, agent.WithCorrectionFallbackAgent(agent.NewMockCorrectionAgent()))
}

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

	return agent.NewLLMScoringAgent(client, agent.WithScoringFallbackAgent(agent.NewMockScoringAgent()))
}

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

	return agent.NewLLMSummaryAgent(client, agent.WithSummaryFallbackAgent(agent.NewMockSummaryAgent()))
}
