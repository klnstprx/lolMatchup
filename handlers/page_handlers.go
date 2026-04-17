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
	Logger          *log.Logger
	ChampionHandler *ChampionHandler
	PlayerHandler   *PlayerHandler
}

// NewPageHandler constructs a PageHandler.
func NewPageHandler(cfg *config.AppConfig, ch *ChampionHandler, ph *PlayerHandler) *PageHandler {
	return &PageHandler{
		Logger:          cfg.Logger,
		ChampionHandler: ch,
		PlayerHandler:   ph,
	}
}

// HomePageGET renders the landing home page.
func (p *PageHandler) HomePageGET(c *gin.Context) {
	p.Logger.Debug("Rendering home page")
	cmp := components.HomePage()
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, cmp))
}

// SearchGET routes a unified search query to the appropriate handler.
// For HTMX requests, it proxies directly to the champion or player handler
// so the result renders inline on the homepage.
// For plain HTTP requests, it redirects to the appropriate page.
func (p *PageHandler) SearchGET(c *gin.Context) {
	// Parse q manually from the URL to avoid initializing Gin's queryCache,
	// which cannot be reset before proxying to another handler.
	q := strings.TrimSpace(c.Request.URL.Query().Get("q"))
	if q == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	isPlayer := strings.Contains(q, "#")

	// HTMX request: proxy to the right handler for inline swap
	if c.GetHeader("HX-Request") == "true" {
		if isPlayer {
			c.Request.URL.RawQuery = "riotID=" + url.QueryEscape(q)
			p.PlayerHandler.PlayerGET(c)
		} else {
			c.Request.URL.RawQuery = "champion=" + url.QueryEscape(q)
			p.ChampionHandler.ChampionGET(c)
		}
		return
	}

	// Non-HTMX: redirect to canonical routes
	var target string
	if isPlayer {
		target = "/player?riotID=" + url.QueryEscape(q)
	} else {
		target = "/champion?champion=" + url.QueryEscape(q)
	}
	c.Redirect(http.StatusFound, target)
}
