package router

import (
	"github.com/gin-gonic/gin"

	"speakmate/internal/handler"
	"speakmate/internal/repository"
	"speakmate/internal/service"
)

// New 创建并配置 Gin 路由引擎。
func New() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/health", handler.Health)

	// Scenario API 当前使用内存仓库，后续可以替换成数据库仓库。
	scenarioRepo := repository.NewMemoryScenarioRepository()
	scenarioService := service.NewScenarioService(scenarioRepo)
	scenarioHandler := handler.NewScenarioHandler(scenarioService)
	sessionRepo := repository.NewMemorySessionRepository()
	sessionService := service.NewSessionService(scenarioService, sessionRepo)
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
