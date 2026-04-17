package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/renderer"
)

type ChampionHandler struct {
	Logger *log.Logger
	Cache  *cache.Cache
	Client *client.Client
	Config *config.AppConfig
}

func NewChampionHandler(cfg *config.AppConfig, client *client.Client) *ChampionHandler {
	return &ChampionHandler{
		Cache:  cfg.Cache,
		Logger: cfg.Logger,
		Config: cfg,
		Client: client,
	}
}

// ChampionGET handles /champion requests with content negotiation.
// HTMX requests get a fragment (with modal/detail/compact variants).
// Full page requests get the champion search page with results pre-populated.
func (h *ChampionHandler) ChampionGET(c *gin.Context) {
	ctx := c.Request.Context()
	inputName := c.Query("champion")
	isHTMX := c.GetHeader("HX-Request") == "true"

	// No query param: show empty search page or error for HTMX
	if inputName == "" {
		if isHTMX {
			renderError(c, http.StatusBadRequest, "Champion name is required.")
			h.Logger.Error("Champion name query param is missing", "url", c.Request.URL.String())
			return
		}
		cmp := components.ChampionPage("", nil)
		c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
		return
	}

	// Lookup champion
	result := h.lookupChampion(ctx, inputName)

	// HTMX: return fragment
	if isHTMX {
		if result.Error != "" {
			renderError(c, http.StatusOK, result.Error)
			return
		}
		// Render based on query param: modal, detail, compact, or default
		var comp templ.Component
		if _, ok := c.GetQuery("modal"); ok {
			comp = components.ChampionModal(result.Champion, result.Config)
		} else if _, ok := c.GetQuery("detail"); ok {
			comp = components.ChampionDetail(result.Champion, result.Config)
		} else if _, ok := c.GetQuery("compact"); ok {
			comp = components.ChampionCompact(result.Champion, result.Config)
		} else {
			comp = components.ChampionComponent(result.Champion, result.Config)
		}
		c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, comp))
		return
	}

	// Full page: wrap in ChampionPage layout
	cmp := components.ChampionPage(inputName, result)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}

// lookupChampion performs a champion lookup for server-side rendering.
// On failure, returns a ChampionResult with the Error field set.
func (h *ChampionHandler) lookupChampion(ctx context.Context, inputName string) *components.ChampionResult {
	championID, err := h.Cache.SearchChampionName(inputName)
	if err != nil {
		h.Logger.Debug("champion lookup: name not found in cache", "error", err)
		championID = inputName
	}

	champion, inCache := h.Cache.GetChampionByID(championID)
	if inCache {
		return &components.ChampionResult{Champion: champion, Config: h.Config}
	}

	fetchedChampion, fetchErr := h.Client.FetchChampionData(ctx, championID)
	if fetchErr != nil {
		h.Logger.Debug("champion lookup: fetch error", "input", inputName, "error", fetchErr)
		if errors.Is(fetchErr, client.ErrChampionNotFound) {
			return &components.ChampionResult{Error: fmt.Sprintf("Champion '%s' not found.", inputName)}
		}
		return &components.ChampionResult{Error: "Error fetching champion data."}
	}
	h.Cache.SetChampion(fetchedChampion)
	return &components.ChampionResult{Champion: fetchedChampion, Config: h.Config}
}
