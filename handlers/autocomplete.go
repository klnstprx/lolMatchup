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

type AutocompleteHandler struct {
	Logger *log.Logger
	Cache  *cache.Cache
	Config *config.AppConfig
	Client *client.Client
}

func NewAutocompleteHandler(cfg *config.AppConfig, apiClient *client.Client) *AutocompleteHandler {
	return &AutocompleteHandler{
		Logger: cfg.Logger,
		Cache:  cfg.Cache,
		Config: cfg,
		Client: apiClient,
	}
}

// AutocompleteGET handles /autocomplete requests for champion names.
func (h *AutocompleteHandler) AutocompleteGET(c *gin.Context) {
	userQuery := strings.TrimSpace(c.Query("champion"))
	// Get up to 10 best fuzzy/autocomplete matches
	var suggestions []string
	if userQuery != "" {
		suggestions = h.Cache.Autocomplete(userQuery, 10)
	}
	comp := components.ChampionAutocomplete(suggestions)
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, comp))
}
