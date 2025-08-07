package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/a-h/templ"
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

func NewChampionHandler(cfg *config.AppConfig, apiClient *client.Client) *ChampionHandler {
	return &ChampionHandler{
		Cache:  cfg.Cache,
		Logger: cfg.Logger,
		Config: cfg,
		Client: apiClient,
	}
}

// ChampionGET handles /champion GET requests, returning champion data in templ format.
func (h *ChampionHandler) ChampionGET(c *gin.Context) {
	ctx := c.Request.Context()
	inputName, ok := c.GetQuery("champion")
	if !ok || inputName == "" {
		c.String(http.StatusBadRequest, "Champion name is required.")
		h.Logger.Error("Champion name query param is missing", "url", c.Request.URL.String())
		return
	}

	championID, err := h.Cache.SearchChampionName(inputName)
	if err != nil {
		h.Logger.Debug("Champion name not found in cache by SearchChampionName", "error", err)
		// We'll try to use the raw inputName as champion ID if no match was found.
		championID = inputName
	}

	champion, inCache := h.Cache.GetChampionByID(championID)
	if inCache {
		h.Logger.Debug("Champion data loaded from cache", "championID", championID)
	} else {
		fetchedChampion, fetchErr := h.Client.FetchChampionData(ctx, championID, h.Config.DDragonURLData)
		if fetchErr != nil {
			if errors.Is(fetchErr, client.ErrChampionNotFound) {
				h.Logger.Debug("Champion not found", "input", inputName, "championID", championID)
				c.String(http.StatusNotFound, fmt.Sprintf("Champion '%s' not found.", inputName))
				return
			}
			h.Logger.Error("Failed to fetch champion detail", "error", fetchErr, "championID", championID)
			c.String(http.StatusInternalServerError, "Error fetching champion data.")
			return
		}
		h.Cache.SetChampion(fetchedChampion)
		champion = fetchedChampion
		h.Logger.Debug("Fetched champion data from DDragon", "championID", champion.ID)
	}

   // Render based on 'modal' or 'compact' query param
   var comp templ.Component
   if _, ok := c.GetQuery("modal"); ok {
       // full component inside modal overlay
       comp = components.ChampionModal(champion, h.Config)
   } else if _, ok := c.GetQuery("compact"); ok {
       // inline compact view
       comp = components.ChampionCompact(champion, h.Config)
   } else {
       // default full component in place
       comp = components.ChampionComponent(champion, h.Config)
   }
   c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, comp))
}
