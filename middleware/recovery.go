package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/config"
)

// RecoveryMiddleware recovers from panics and logs errors using charmbracelet/log
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				config.App.Logger.Error("Recovered from panic in HTTP request",
					"error", err,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
				)

				// Respond with 500 Internal Server Error
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		// Continue to next middleware/handler
		c.Next()
	}
}
