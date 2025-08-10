package middleware

import (
	"strconv"
	"time"

	"task/internal/metrics"

	"github.com/gin-gonic/gin"
)

func HTTPMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		labels := []string{c.Request.Method, c.FullPath(), strconv.Itoa(status)}
		metrics.HttpRequestsTotal.WithLabelValues(labels...).Inc()
		metrics.HttpRequestDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
	}
}
