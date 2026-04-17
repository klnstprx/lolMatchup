package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/renderer"
)

// renderError renders an ErrorMessage component with the given HTTP status code.
func renderError(c *gin.Context, status int, msg string) {
	ctx := c.Request.Context()
	c.Render(status, renderer.New(ctx, status, components.ErrorMessage(msg)))
}
