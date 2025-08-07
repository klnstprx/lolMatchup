package handlers

import (
	"net/http"

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

// ChampionPageGET renders the champion lookup page.
func (p *PageHandler) ChampionPageGET(c *gin.Context) {
	p.Logger.Debug("Rendering champion lookup page")
	cmp := components.ChampionPage()
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, cmp))
}

// HomePageGET renders the landing home page.
func (p *PageHandler) HomePageGET(c *gin.Context) {
	p.Logger.Debug("Rendering home page")
	cmp := components.HomePage()
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, cmp))
}

// PlayerPageGET renders the player lookup page.
func (p *PageHandler) PlayerPageGET(c *gin.Context) {
	p.Logger.Debug("Rendering player lookup page")
	cmp := components.PlayerPage()
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, cmp))
}

// LiveGamePageGET renders the live game lookup page.
func (p *PageHandler) LiveGamePageGET(c *gin.Context) {
	p.Logger.Debug("Rendering live game lookup page")
	cmp := components.LiveGamePage()
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, cmp))
}
