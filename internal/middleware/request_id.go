package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderRequestID   = "X-Request-ID"
	ContextKeyTraceID = "trace_id"
)

// RequestID ensures every request has a trace_id (request id) available.
// - reads from X-Request-ID if present; otherwise generates a UUID v4
// - sets header X-Request-ID on the response
// - stores into gin.Context with key "trace_id"
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Header(HeaderRequestID, rid)
		c.Set(ContextKeyTraceID, rid)
		c.Next()
	}
}

// TraceIDFromContext returns trace_id string if set.
func TraceIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(ContextKeyTraceID); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}
