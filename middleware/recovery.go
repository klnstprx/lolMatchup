package middleware

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware recovers from panics and logs errors using charmbracelet/log
func RecoveryMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Recovered from panic in HTTP request",
					"error", err,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
				)

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
