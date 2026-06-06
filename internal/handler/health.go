package handler

import (
	"github.com/gin-gonic/gin"

	"speakmate/internal/response"
)

func Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "ok",
	})
}
