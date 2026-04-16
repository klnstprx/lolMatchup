package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/renderer"
)

// LiveGameHandler handles live game search requests.
type LiveGameHandler struct {
	Logger *log.Logger
	Client *client.Client
	Config *config.AppConfig
}

// NewLiveGameHandler creates a LiveGameHandler.
func NewLiveGameHandler(cfg *config.AppConfig, apiClient *client.Client) *LiveGameHandler {
	return &LiveGameHandler{
		Logger: cfg.Logger,
		Client: apiClient,
		Config: cfg,
	}
}

// LiveGameGET handles GET /livegame requests.
func (h *LiveGameHandler) LiveGameGET(c *gin.Context) {
	// Validate Riot ID (gameName + tagLine)
	ctx := c.Request.Context()
	// Parse Riot ID in form nickname#tag
	riotID, ok := c.GetQuery("riotID")
	if !ok || riotID == "" {
		c.String(http.StatusBadRequest, "Summoner identifier is required (nickname#tag).")
		h.Logger.Error("riotID query param missing", "url", c.Request.URL.String())
		return
	}
	idParts := strings.SplitN(riotID, "#", 2)
	if len(idParts) != 2 {
		c.String(http.StatusBadRequest, "Invalid format for Summoner; use nickname#tag.")
		h.Logger.Error("invalid riotID format", "riotID", riotID)
		return
	}
	gameName, tagLine := idParts[0], idParts[1]

	// Step 1: Fetch encrypted PUUID via account-v1
	acct, err := h.Client.FetchAccountByRiotID(ctx, gameName, tagLine, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		switch {
		case errors.Is(err, client.ErrAccountNotFound):
			c.String(http.StatusNotFound, fmt.Sprintf("Account '%s#%s' not found.", gameName, tagLine))
			h.Logger.Debug("Account not found", "riotID", gameName+"#"+tagLine)
			return
		case errors.Is(err, client.ErrPermissionDenied):
			c.String(http.StatusForbidden, "Permission denied: check your Riot API key and region.")
			h.Logger.Error("Permission denied fetching account info", "error", err)
			return
		default:
			h.Logger.Error("Error fetching account info", "error", err)
			c.String(http.StatusInternalServerError, "Error fetching account data.")
			return
		}
	}

	// Step 2: Fetch current game via spectator-v5 using PUUID
	activeGame, err := h.Client.FetchCurrentGameByPUUID(ctx, acct.PUUID, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		switch {
		case errors.Is(err, client.ErrGameNotFound):
			c.String(http.StatusNotFound, fmt.Sprintf("Account '%s#%s' is not currently in a game.", gameName, tagLine))
			h.Logger.Debug("No active game for account", "riotID", gameName+"#"+tagLine)
			return
		case errors.Is(err, client.ErrPermissionDenied):
			c.String(http.StatusForbidden, "Permission denied: check your Riot API key and region.")
			h.Logger.Error("Permission denied fetching active game", "error", err)
			return
		default:
			h.Logger.Error("Error fetching active game", "error", err)
			c.String(http.StatusInternalServerError, "Error fetching live game data.")
			return
		}
	}

	// Find the user's team
	var userTeamID int64
	for _, p := range activeGame.Participants {
		if p.RiotID == riotID {
			userTeamID = p.TeamID
			break
		}
	}

	// Build opponent list: lookup numeric championId -> textual ID -> display name
	keyMap := h.Config.Cache.GetChampionKeyMap() // numeric key -> textual ID
	nameMap := h.Config.Cache.GetChampionMap()   // champion Name -> textual ID
	textualToName := make(map[string]string, len(nameMap))
	for name, id := range nameMap {
		textualToName[id] = name
	}

	var parts []map[string]string
	for _, p := range activeGame.Participants {
		if p.TeamID == userTeamID {
			continue
		}
		numKey := strconv.FormatInt(p.ChampionID, 10)
		textID, found := keyMap[numKey]
		if !found {
			h.Logger.Error("Champion ID missing from keyMap", "championId", numKey)
			textID = numKey
		}
		champName, foundName := textualToName[textID]
		if !foundName {
			champName = textID
		}
		parts = append(parts, map[string]string{
			"riotId":       p.RiotID,
			"championId":   textID,
			"championName": champName,
		})
	}

	// Render using templ component
	cmp := components.LiveGameInfo(parts, h.Config)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}
