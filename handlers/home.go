package handlers

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/renderer"
)

type HomeHandler struct {
	Logger *log.Logger
	Cache  *cache.Cache
	Client *client.Client
	Config *config.AppConfig
}

func NewHomeHandler(cfg *config.AppConfig, apiClient *client.Client) *HomeHandler {
	return &HomeHandler{
		Cache:  cfg.Cache,
		Logger: cfg.Logger,
		Config: cfg,
		Client: apiClient,
	}
}

func (h *HomeHandler) HomeGET(c *gin.Context) {
	r := renderer.New(c.Request.Context(), http.StatusOK, components.Home())
	c.Render(http.StatusOK, r)
}
