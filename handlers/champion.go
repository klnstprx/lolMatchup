package handlers

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
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

func (h *ChampionHandler) ChampionGET(c *gin.Context) {
	ctx := c.Request.Context()
	inputName, ok := c.GetQuery("champion")
	if !ok || len(inputName) == 0 {
		c.String(http.StatusBadRequest, "Champion name is required.")
		h.Logger.Error("Champion name empty!", "url", c.Request.URL)
		return
	}

	var champion models.Champion
	cachedChampionID, err := h.Cache.SearchChampionName(inputName)
	if err != nil {
		h.Logger.Debug("Champion's name not found in cache.", "error", err)
	}

	var inCache bool
	if cachedChampionID != "" {
		// Champion's name found in cache, check if data is already cached
		champion, inCache = h.Cache.GetChampionByID(cachedChampionID)
		if inCache {
			h.Logger.Debug("Champion data found in cache.", "championID", cachedChampionID)
		}
	}

	if !inCache {
		championID := cachedChampionID
		if championID == "" {
			championID = inputName
		}

		champion, err = h.Client.FetchChampionData(ctx, championID, h.Config.DDragonURLData)
		h.Logger.Debug("Fetching champion data.")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching champion data.")
			h.Logger.Error("Error fetching champion data.", "error", err)
			return
		}

		// Cache the champion data
		h.Logger.Debug("Caching champion data.", "championID", champion.ID)
		h.Cache.SetChampion(champion)
	}

	h.Logger.Debug("Serving champion data.", "championID", champion.ID)

	r := renderer.New(c.Request.Context(), http.StatusOK, components.ChampionComponent(champion, h.Config))
	c.Render(http.StatusOK, r)
}
