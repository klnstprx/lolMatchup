package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/config"
)

// LoggerMiddleware logs HTTP requests using charmbracelet/log
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		latency := time.Since(start)

		// Get status code
		status := c.Writer.Status()

		// Log details
		config.App.Logger.Info("Incoming request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", latency,
			"clientIP", c.ClientIP(),
			"userAgent", c.Request.UserAgent(),
		)
	}
}
