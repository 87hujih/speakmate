package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"speakmate/internal/config"
	"speakmate/internal/response"
	"speakmate/internal/security"
)

// 当前模块使用的常量。
const (
	internalServerErrorCode = 9001
	requestTimeoutCode      = 9002
	requestBodyTooLargeCode = 9003
	rateLimitExceededCode   = 9004
)

// rateLimitBucket 保存单个客户端在当前限流窗口内的计数。
type rateLimitBucket struct {
	count   int
	resetAt time.Time
}

// CORS 应用跨域访问配置，并处理浏览器预检请求。
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	allowWildcard := false
	for _, origin := range cfg.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowWildcard = true
			continue
		}
		allowedOrigins[origin] = struct{}{}
	}

	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			switch {
			case allowWildcard && !cfg.AllowCredentials:
				c.Header("Access-Control-Allow-Origin", "*")
			case allowWildcard:
				c.Header("Access-Control-Allow-Origin", origin)
			default:
				if _, ok := allowedOrigins[origin]; ok {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}

			if c.Writer.Header().Get("Access-Control-Allow-Origin") != "" {
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Methods", allowedMethods)
				c.Header("Access-Control-Allow-Headers", allowedHeaders)
				if cfg.AllowCredentials {
					c.Header("Access-Control-Allow-Credentials", "true")
				}
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLogger 为每个请求写入一行已脱敏的访问日志。
func RequestLogger(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.Default()
	}

	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		logger.Printf(
			"请求 方法=%s 路径=%s 状态=%d 耗时=%s",
			c.Request.Method,
			security.RedactURL(c.Request.URL.RequestURI()),
			c.Writer.Status(),
			time.Since(startedAt).Round(time.Millisecond),
		)
	}
}

// BodySizeLimit 在 Handler 解析 JSON 或 multipart 前限制请求体大小。
// 音频接口仍保留更严格的文件级校验。
func BodySizeLimit(maxBytes int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 || c.Request == nil || c.Request.Body == nil {
			c.Next()
			return
		}

		if c.Request.ContentLength > int64(maxBytes) {
			c.Abort()
			response.Error(c, http.StatusRequestEntityTooLarge, requestBodyTooLargeCode, "请求体过大")
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxBytes))
		c.Next()
	}
}

// RateLimit 提供按客户端 IP 计数的内存限流，用于 Demo 和基础部署防护。
// 它不承担分布式生产配额能力。
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]rateLimitBucket)

	return func(c *gin.Context) {
		if requests <= 0 || window <= 0 || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		now := time.Now()
		key := c.ClientIP()
		if key == "" {
			key = "unknown"
		}

		mu.Lock()
		bucket := buckets[key]
		if bucket.resetAt.IsZero() || now.After(bucket.resetAt) {
			bucket = rateLimitBucket{resetAt: now.Add(window)}
		}
		if bucket.count >= requests {
			mu.Unlock()
			c.Abort()
			response.Error(c, http.StatusTooManyRequests, rateLimitExceededCode, "请求过于频繁")
			return
		}
		bucket.count++
		buckets[key] = bucket
		mu.Unlock()

		c.Next()
	}
}

// Recover 捕获 panic 并返回统一 API 错误结构，避免向客户端暴露 panic 细节。
func Recover(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.Default()
	}

	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Printf(
					"已捕获 panic 方法=%s 路径=%s 错误=%s",
					c.Request.Method,
					security.RedactURL(c.Request.URL.RequestURI()),
					security.RedactString(fmt.Sprint(recovered)),
				)
				c.Abort()
				if !c.Writer.Written() {
					response.Error(c, http.StatusInternalServerError, internalServerErrorCode, "服务器内部错误")
				}
			}
		}()

		c.Next()
	}
}

// RequestTimeout 为请求上下文设置截止时间，超时后会取消使用该上下文的下游调用。
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if timeout <= 0 || isLongLivedRequest(c.Request.URL.Path) {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() == context.DeadlineExceeded && !c.Writer.Written() {
			c.Abort()
			response.Error(c, http.StatusGatewayTimeout, requestTimeoutCode, "请求超时")
		}
	}
}

// isLongLivedRequest 判断请求是否属于长连接链路。
func isLongLivedRequest(path string) bool {
	return strings.HasSuffix(path, "/stream") || strings.HasSuffix(path, "/audio/ws")
}
