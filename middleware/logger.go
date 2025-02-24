package middleware

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

// Logging gin requests.
func LoggerMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)

		status := c.Writer.Status()

		logger.Info("Incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", latency,
			"clientIP", c.ClientIP(),
			"userAgent", c.Request.UserAgent(),
		)
	}
}
