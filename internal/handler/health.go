package handler

import (
	"github.com/gin-gonic/gin"

	"speakmate/internal/response"
)

// Health 返回服务健康检查结果。
func Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}
