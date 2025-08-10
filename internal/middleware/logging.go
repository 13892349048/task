package middleware

import (
	"time"

	"task/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger logs method, path, status, latency, client IP, and optional user_id.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		clientIP := c.ClientIP()
		traceID := TraceIDFromContext(c)

		var userID any
		if v, ok := c.Get("user_id"); ok {
			userID = v
		}

		logger.L().Info("request completed",
			zap.String("trace_id", traceID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency_ms", latency),
			zap.String("client_ip", clientIP),
			zap.Any("user_id", userID),
		)
	}
}
