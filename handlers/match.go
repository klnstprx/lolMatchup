package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
	"github.com/klnstprx/lolMatchup/renderer"
)

// MatchHandler handles match detail requests.
type MatchHandler struct {
	Logger *log.Logger
	Client *client.Client
	Config *config.AppConfig
}

// NewMatchHandler constructs a MatchHandler.
func NewMatchHandler(cfg *config.AppConfig, apiClient *client.Client) *MatchHandler {
	return &MatchHandler{Logger: cfg.Logger, Client: apiClient, Config: cfg}
}

// MatchGET handles GET /match?id=...&puuid=... requests.
func (h *MatchHandler) MatchGET(c *gin.Context) {
	ctx := c.Request.Context()

	matchID := c.Query("id")
	if matchID == "" {
		renderError(c, http.StatusBadRequest, "Match ID is required.")
		return
	}

	puuid := c.Query("puuid")

	match, err := h.Client.FetchMatch(ctx, matchID, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		if errors.Is(err, client.ErrMatchNotFound) {
			renderError(c, http.StatusNotFound, fmt.Sprintf("Match '%s' not found.", matchID))
			h.Logger.Debug("match not found", "matchId", matchID)
			return
		}
		h.Logger.Error("failed to fetch match", "matchId", matchID, "error", err)
		renderError(c, http.StatusInternalServerError, "Error fetching match data.")
		return
	}

	cmp := components.MatchDetailView(match, puuid, h.Config)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}

// MatchPlayerGET handles GET /match/player?id=...&puuid=... — renders a detailed stats modal for one player.
func (h *MatchHandler) MatchPlayerGET(c *gin.Context) {
	ctx := c.Request.Context()

	matchID := c.Query("id")
	puuid := c.Query("puuid")
	if matchID == "" || puuid == "" {
		renderError(c, http.StatusBadRequest, "Match ID and player PUUID are required.")
		return
	}

	match, err := h.Client.FetchMatch(ctx, matchID, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		if errors.Is(err, client.ErrMatchNotFound) {
			renderError(c, http.StatusNotFound, fmt.Sprintf("Match '%s' not found.", matchID))
			h.Logger.Debug("match not found", "matchId", matchID)
			return
		}
		h.Logger.Error("failed to fetch match", "matchId", matchID, "error", err)
		renderError(c, http.StatusInternalServerError, "Error fetching match data.")
		return
	}

	statsCtx := buildPlayerStatsContext(match, puuid, h.Config.PatchNumber)
	if statsCtx == nil {
		renderError(c, http.StatusNotFound, "Player not found in this match.")
		return
	}

	cmp := components.PlayerStatsModal(*statsCtx)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}

func buildPlayerStatsContext(match models.MatchDTO, puuid, patchNumber string) *models.PlayerStatsContext {
	var player *models.MatchParticipant
	var teamKills int
	var teamTotalDamage int
	var maxGold, maxVision, maxDamage int

	for i, p := range match.Info.Participants {
		if p.PUUID == puuid {
			player = &match.Info.Participants[i]
		}
		if p.GoldEarned > maxGold {
			maxGold = p.GoldEarned
		}
		if p.VisionScore > maxVision {
			maxVision = p.VisionScore
		}
		if p.TotalDamageDealtToChampions > maxDamage {
			maxDamage = p.TotalDamageDealtToChampions
		}
	}
	if player == nil {
		return nil
	}

	for _, p := range match.Info.Participants {
		if p.TeamID == player.TeamID {
			teamKills += p.Kills
			teamTotalDamage += p.TotalDamageDealtToChampions
		}
	}

	var kp float64
	if teamKills > 0 {
		kp = float64(player.Kills+player.Assists) / float64(teamKills)
	}

	return &models.PlayerStatsContext{
		Player:            *player,
		MatchID:           match.Metadata.MatchID,
		GameDuration:      match.Info.GameDuration,
		TeamTotalDamage:   teamTotalDamage,
		MaxGoldInGame:     maxGold,
		MaxVisionInGame:   maxVision,
		MaxDamageInGame:   maxDamage,
		KillParticipation: kp,
		PatchNumber:       patchNumber,
	}
}
