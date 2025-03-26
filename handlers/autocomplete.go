package handlers

import (
	"net/http"
	"sort"
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
	if userQuery == "" {
		empty := components.ChampionAutocomplete([]string{})
		c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, empty))
		return
	}

	queryLower := strings.ToLower(userQuery)
	championMap := h.Cache.GetChampionMap()

	// Gather all champion names in a slice for alphabetical sorting
	allNames := make([]string, 0, len(championMap))
	for name := range championMap {
		allNames = append(allNames, name)
	}

	sort.Strings(allNames) // alphabetical order

	var prefixMatches, substringMatches []string

	for _, name := range allNames {
		nameLower := strings.ToLower(name)
		if strings.HasPrefix(nameLower, queryLower) {
			prefixMatches = append(prefixMatches, name)
		} else if strings.Contains(nameLower, queryLower) {
			substringMatches = append(substringMatches, name)
		}
	}

	// Combine prefix matches (first) and substring matches (second)
	results := append(prefixMatches, substringMatches...)
	// Trim to a maximum of 10
	if len(results) > 10 {
		results = results[:10]
	}

	autocomplete := components.ChampionAutocomplete(results)
	c.Render(http.StatusOK, renderer.New(c.Request.Context(), http.StatusOK, autocomplete))
}
