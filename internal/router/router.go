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
	sessionService := service.NewSessionService(
		scenarioService,
		sessionRepo,
		service.WithConversationAgent(NewConversationAgent(cfg)),
	)
	sessionHandler := handler.NewSessionHandler(sessionService)
	messageHandler := handler.NewMessageHandler(sessionService)

	// v1 API 路由组承载场景、训练 Session 和消息等后续接口。
	api := engine.Group("/api/v1")
	api.GET("/scenarios", scenarioHandler.List)
	api.GET("/scenarios/:id", scenarioHandler.Detail)
	api.POST("/sessions", sessionHandler.Create)
	api.GET("/sessions/:id", sessionHandler.Detail)
	api.POST("/sessions/:id/finish", sessionHandler.Finish)
	api.POST("/sessions/:id/messages", messageHandler.Send)

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
