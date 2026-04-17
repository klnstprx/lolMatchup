package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/renderer"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware returns a Gin middleware that limits requests using a
// token bucket. rps is requests per second; burst is the maximum burst size.
// This is a global (not per-IP) limiter, suitable for protecting upstream API
// quotas such as Riot API rate limits.
func RateLimitMiddleware(rps rate.Limit, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rps, burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			ctx := c.Request.Context()
			c.Render(http.StatusTooManyRequests, renderer.New(ctx, http.StatusTooManyRequests, components.ErrorMessage("Rate limit exceeded. Please try again shortly.")))
			c.Abort()
			return
		}
		c.Next()
	}
}
