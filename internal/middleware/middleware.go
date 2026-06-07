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

const (
	internalServerErrorCode = 9001
	requestTimeoutCode      = 9002
	requestBodyTooLargeCode = 9003
	rateLimitExceededCode   = 9004
)

type rateLimitBucket struct {
	count   int
	resetAt time.Time
}

// CORS applies the configured cross-origin policy and handles preflight requests.
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

// RequestLogger writes one redacted access log line per request.
func RequestLogger(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.Default()
	}

	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		logger.Printf(
			"request method=%s path=%s status=%d duration=%s",
			c.Request.Method,
			security.RedactURL(c.Request.URL.RequestURI()),
			c.Writer.Status(),
			time.Since(startedAt).Round(time.Millisecond),
		)
	}
}

// BodySizeLimit caps request bodies before handlers parse JSON or multipart
// payloads. Audio endpoints still keep their stricter file-level validation.
func BodySizeLimit(maxBytes int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 || c.Request == nil || c.Request.Body == nil {
			c.Next()
			return
		}

		if c.Request.ContentLength > int64(maxBytes) {
			c.Abort()
			response.Error(c, http.StatusRequestEntityTooLarge, requestBodyTooLargeCode, "request body too large")
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxBytes))
		c.Next()
	}
}

// RateLimit provides a simple in-memory per-client-IP abuse guard. It is meant
// for local demo and basic deployment hygiene, not distributed production quota.
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
			response.Error(c, http.StatusTooManyRequests, rateLimitExceededCode, "rate limit exceeded")
			return
		}
		bucket.count++
		buckets[key] = bucket
		mu.Unlock()

		c.Next()
	}
}

// Recover catches panics and returns the standard API error shape without
// exposing panic details to the client.
func Recover(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.Default()
	}

	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Printf(
					"panic recovered method=%s path=%s error=%s",
					c.Request.Method,
					security.RedactURL(c.Request.URL.RequestURI()),
					security.RedactString(fmt.Sprint(recovered)),
				)
				c.Abort()
				if !c.Writer.Written() {
					response.Error(c, http.StatusInternalServerError, internalServerErrorCode, "internal server error")
				}
			}
		}()

		c.Next()
	}
}

// RequestTimeout attaches a deadline to the request context. Handlers and
// outbound calls using c.Request.Context() will be cancelled when it expires.
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
			response.Error(c, http.StatusGatewayTimeout, requestTimeoutCode, "request timeout")
		}
	}
}

func isLongLivedRequest(path string) bool {
	return strings.HasSuffix(path, "/stream") || strings.HasSuffix(path, "/audio/ws")
}
