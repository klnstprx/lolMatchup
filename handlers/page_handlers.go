package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/renderer"
)

// PageHandler handles rendering of standalone lookup pages.
type PageHandler struct {
	Logger *log.Logger
}

// NewPageHandler constructs a PageHandler.
func NewPageHandler(cfg *config.AppConfig) *PageHandler {
	return &PageHandler{Logger: cfg.Logger}
}

// HomePageGET renders the landing home page.
func (p *PageHandler) HomePageGET(c *gin.Context) {
	p.Logger.Debug("Rendering home page")
	cmp := components.HomePage()
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, cmp))
}

// SearchGET routes a unified search query to the appropriate lookup page.
// If the query contains "#", it redirects to player search; otherwise to champion search.
// For HTMX requests it uses the HX-Redirect header; for plain requests a 302 redirect.
func (p *PageHandler) SearchGET(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}
	var target string
	if strings.Contains(q, "#") {
		target = "/player-search?riotID=" + url.QueryEscape(q)
	} else {
		target = "/champion-search?champion=" + url.QueryEscape(q)
	}
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", target)
		c.Status(http.StatusOK)
		return
	}
	c.Redirect(http.StatusFound, target)
}
