package middleware

import "github.com/gin-gonic/gin"

// CacheControl returns a middleware that sets the Cache-Control header.
func CacheControl(value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", value)
		c.Next()
	}
}
