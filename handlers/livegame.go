package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
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
	ctx := c.Request.Context()
	summonerName, ok := c.GetQuery("summoner")
	if !ok || summonerName == "" {
		c.String(http.StatusBadRequest, "Summoner name is required.")
		h.Logger.Error("Summoner name query param missing", "url", c.Request.URL.String())
		return
	}

	summoner, err := h.Client.FetchSummonerByName(ctx, summonerName, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		if err == client.ErrSummonerNotFound {
			c.String(http.StatusNotFound, fmt.Sprintf("Summoner '%s' not found.", summonerName))
			h.Logger.Debug("Summoner not found", "name", summonerName)
			return
		}
		h.Logger.Error("Error fetching summoner info", "error", err)
		c.String(http.StatusInternalServerError, "Error fetching summoner data.")
		return
	}

	activeGame, err := h.Client.FetchActiveGame(ctx, summoner.ID, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		if err == client.ErrGameNotFound {
			c.String(http.StatusNotFound, fmt.Sprintf("Summoner '%s' is not currently in a game.", summonerName))
			h.Logger.Debug("No active game for summoner", "name", summonerName)
			return
		}
		h.Logger.Error("Error fetching active game", "error", err)
		c.String(http.StatusInternalServerError, "Error fetching live game data.")
		return
	}

	// Pretty-print JSON response inside a <pre> tag.
	jsonData, err := json.MarshalIndent(activeGame, "", "  ")
	if err != nil {
		h.Logger.Error("Error marshalling active game JSON", "error", err)
		c.String(http.StatusInternalServerError, "Error processing live game data.")
		return
	}
	escaped := template.HTMLEscapeString(string(jsonData))
	html := fmt.Sprintf("<pre>%s</pre>", escaped)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
