package middleware

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs requests with timing, status, client IP, etc.
func LoggerMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("Incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", latency.String(),
			"clientIP", c.ClientIP(),
			"userAgent", c.Request.UserAgent(),
		)
	}
}
