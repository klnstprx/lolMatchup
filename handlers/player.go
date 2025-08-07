package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/renderer"
)

// PlayerHandler handles player lookup requests.
type PlayerHandler struct {
	Logger *log.Logger
	Client *client.Client
	Config *config.AppConfig
}

// NewPlayerHandler constructs a PlayerHandler.
func NewPlayerHandler(cfg *config.AppConfig, apiClient *client.Client) *PlayerHandler {
	return &PlayerHandler{Logger: cfg.Logger, Client: apiClient, Config: cfg}
}

// PlayerGET handles /player requests, returning player info via templ component.
func (h *PlayerHandler) PlayerGET(c *gin.Context) {
	ctx := c.Request.Context()
	riotID, ok := c.GetQuery("riotID")
	if !ok || riotID == "" {
		c.String(http.StatusBadRequest, "Summoner identifier is required (nickname#tag).")
		h.Logger.Error("riotID query param missing", "url", c.Request.URL.String())
		return
	}
	parts := strings.SplitN(riotID, "#", 2)
	if len(parts) != 2 {
		c.String(http.StatusBadRequest, "Invalid format for Summoner; use nickname#tag.")
		h.Logger.Error("invalid riotID format", "riotID", riotID)
		return
	}
	player, err := h.Client.FetchSummonerByName(ctx, parts[0], h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		if err == client.ErrSummonerNotFound {
			c.String(http.StatusNotFound, fmt.Sprintf("Summoner '%s' not found.", riotID))
			h.Logger.Debug("summoner not found", "riotID", riotID)
			return
		}
		if errors.Is(err, client.ErrPermissionDenied) {
			c.String(http.StatusForbidden, "Permission denied: check your Riot API key and region.")
			h.Logger.Error("permission denied fetching summoner info", "error", err)
			return
		}
		h.Logger.Error("error fetching summoner info", "error", err)
		c.String(http.StatusInternalServerError, "Error fetching summoner data.")
		return
	}
	cmp := components.PlayerComponent(player)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}
