package router

import (
	"strings"

	"github.com/gin-gonic/gin"

	"speakmate/internal/agent"
	"speakmate/internal/config"
	"speakmate/internal/handler"
	"speakmate/internal/infra/llm"
	"speakmate/internal/repository"
	"speakmate/internal/service"
)

// New 创建并配置 Gin 路由引擎。
func New(configs ...config.Config) *gin.Engine {
	cfg := config.Load()
	if len(configs) > 0 {
		cfg = configs[0]
	}

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/health", handler.Health)

	// Scenario API 当前使用内存仓库，后续可以替换成数据库仓库。
	scenarioRepo := repository.NewMemoryScenarioRepository()
	scenarioService := service.NewScenarioService(scenarioRepo)
	scenarioHandler := handler.NewScenarioHandler(scenarioService)
	sessionRepo := repository.NewMemorySessionRepository()
	feedbackRepo := repository.NewMemoryFeedbackRepository()
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
	reportRepo := repository.NewMemoryReportRepository()
	reportService := service.NewReportService(
		scenarioService,
		sessionRepo,
		feedbackRepo,
		reportRepo,
		service.WithSummaryAgent(NewSummaryAgent(cfg)),
	)
	reportHandler := handler.NewReportHandler(reportService)

	// v1 API 路由组承载场景、训练 Session 和消息等后续接口。
	api := engine.Group("/api/v1")
	api.GET("/scenarios", scenarioHandler.List)
	api.GET("/scenarios/:id", scenarioHandler.Detail)
	api.POST("/sessions", sessionHandler.Create)
	api.GET("/sessions/:id", sessionHandler.Detail)
	api.POST("/sessions/:id/finish", sessionHandler.Finish)
	api.POST("/sessions/:id/messages", messageHandler.Send)
	api.GET("/sessions/:id/corrections", feedbackHandler.ListSessionCorrections)
	api.GET("/sessions/:id/scores", feedbackHandler.GetSessionScore)
	api.POST("/sessions/:id/report", reportHandler.Generate)
	api.GET("/sessions/:id/report", reportHandler.Get)
	api.GET("/messages/:message_id/corrections", feedbackHandler.GetMessageCorrection)

	return engine
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
