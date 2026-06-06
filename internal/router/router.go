package router

import (
	"github.com/gin-gonic/gin"

	"speakmate/internal/handler"
)

func New() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/health", handler.Health)

	return engine
}
