package handlers

import (
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/renderer"
)

const defaultAutocompleteLimit = 10

type AutocompleteHandler struct {
	Logger *log.Logger
	Cache  *cache.Cache
	Config *config.AppConfig
	Client *client.Client
}

func NewAutocompleteHandler(cfg *config.AppConfig, client *client.Client) *AutocompleteHandler {
	return &AutocompleteHandler{
		Logger: cfg.Logger,
		Cache:  cfg.Cache,
		Config: cfg,
		Client: client,
	}
}

// AutocompleteGET handles /autocomplete requests for champion names.
// Accepts "champion" or "q" query param to support both champion form and unified search.
func (h *AutocompleteHandler) AutocompleteGET(c *gin.Context) {
	userQuery := strings.TrimSpace(c.Query("champion"))
	if userQuery == "" {
		userQuery = strings.TrimSpace(c.Query("q"))
	}
	var suggestions []cache.AutocompleteResult
	if userQuery != "" {
		suggestions = h.Cache.AutocompleteRich(userQuery, defaultAutocompleteLimit)
	}
	comp := components.ChampionAutocomplete(suggestions, userQuery, h.Config.PatchNumber)
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, comp))
}
