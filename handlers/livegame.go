package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/components"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
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
		renderError(c, http.StatusBadRequest, "Summoner identifier is required (nickname#tag).")
		h.Logger.Error("riotID query param missing", "url", c.Request.URL.String())
		return
	}
	idParts := strings.SplitN(riotID, "#", 2)
	if len(idParts) != 2 {
		renderError(c, http.StatusBadRequest, "Invalid format for Summoner; use nickname#tag.")
		h.Logger.Error("invalid riotID format", "riotID", riotID)
		return
	}
	gameName, tagLine := idParts[0], idParts[1]

	// Step 1: Fetch encrypted PUUID via account-v1
	acct, err := h.Client.FetchAccountByRiotID(ctx, gameName, tagLine, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		switch {
		case errors.Is(err, client.ErrAccountNotFound):
			renderError(c, http.StatusNotFound, fmt.Sprintf("Account '%s#%s' not found.", gameName, tagLine))
			h.Logger.Debug("Account not found", "riotID", gameName+"#"+tagLine)
			return
		case errors.Is(err, client.ErrPermissionDenied):
			renderError(c, http.StatusForbidden, "Permission denied: check your Riot API key and region.")
			h.Logger.Error("Permission denied fetching account info", "error", err)
			return
		default:
			h.Logger.Error("Error fetching account info", "error", err)
			renderError(c, http.StatusInternalServerError, "Error fetching account data.")
			return
		}
	}

	// Step 2: Fetch current game via spectator-v5 using PUUID
	activeGame, err := h.Client.FetchCurrentGameByPUUID(ctx, acct.PUUID, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		switch {
		case errors.Is(err, client.ErrGameNotFound):
			renderError(c, http.StatusNotFound, fmt.Sprintf("Account '%s#%s' is not currently in a game.", gameName, tagLine))
			h.Logger.Debug("No active game for account", "riotID", gameName+"#"+tagLine)
			return
		case errors.Is(err, client.ErrPermissionDenied):
			renderError(c, http.StatusForbidden, "Permission denied: check your Riot API key and region.")
			h.Logger.Error("Permission denied fetching active game", "error", err)
			return
		default:
			h.Logger.Error("Error fetching active game", "error", err)
			renderError(c, http.StatusInternalServerError, "Error fetching live game data.")
			return
		}
	}

	vd := h.buildViewData(activeGame, riotID)
	if !vd.found {
		h.Logger.Warn("Player not found in participant list", "riotID", riotID)
		renderError(c, http.StatusNotFound, fmt.Sprintf("Could not identify '%s' in the current game's participant list.", riotID))
		return
	}

	// Enrich opponents with recent match data (best-effort, non-blocking)
	h.enrichOpponents(ctx, vd.parts)

	cmp := components.LiveGameInfo(vd.parts, h.Config, riotID, vd.userChampionName, vd.userChampionID)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}

// PlayerLiveGameGET handles GET /player/livegame?puuid=...&riotID=...
// Returns a LiveGameStatus fragment for embedding in the player page.
// Designed to be loaded via HTMX (hx-trigger="load" then "every 30s").
func (h *LiveGameHandler) PlayerLiveGameGET(c *gin.Context) {
	ctx := c.Request.Context()
	puuid := c.Query("puuid")
	riotID := c.Query("riotID")
	if puuid == "" || riotID == "" {
		c.Status(http.StatusNoContent)
		return
	}

	activeGame, err := h.Client.FetchCurrentGameByPUUID(ctx, puuid, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		if errors.Is(err, client.ErrGameNotFound) {
			// Not in game — render status with polling
			cmp := components.LiveGameStatus(false, riotID, nil, h.Config, "", "", puuid, time.Now())
			c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
			return
		}
		// Other errors: graceful degradation, render not-in-game
		h.Logger.Debug("player livegame check failed", "puuid", puuid, "error", err)
		cmp := components.LiveGameStatus(false, riotID, nil, h.Config, "", "", puuid, time.Now())
		c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
		return
	}

	vd := h.buildViewData(activeGame, riotID)
	if !vd.found {
		cmp := components.LiveGameStatus(false, riotID, nil, h.Config, "", "", puuid, time.Now())
		c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
		return
	}

	h.enrichOpponents(ctx, vd.parts)

	cmp := components.LiveGameStatus(true, riotID, vd.parts, h.Config, vd.userChampionName, vd.userChampionID, puuid, time.Now())
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}

// liveGameViewData holds the resolved participant data for rendering.
type liveGameViewData struct {
	parts            []components.OpponentView
	userChampionID   string
	userChampionName string
	found            bool
}

// buildViewData resolves champion IDs and splits participants into user team vs opponents.
func (h *LiveGameHandler) buildViewData(game models.CurrentGameInfo, riotID string) liveGameViewData {
	keyMap := h.Config.Cache.GetChampionKeyMap()
	nameMap := h.Config.Cache.GetChampionMap()
	textualToName := make(map[string]string, len(nameMap))
	for name, id := range nameMap {
		textualToName[id] = name
	}

	resolve := func(championID int64) (textID, champName string) {
		numKey := strconv.FormatInt(championID, 10)
		textID, ok := keyMap[numKey]
		if !ok {
			h.Logger.Error("Champion ID missing from keyMap", "championId", numKey)
			textID = numKey
		}
		champName, ok = textualToName[textID]
		if !ok {
			champName = textID
		}
		return textID, champName
	}

	var vd liveGameViewData
	var userTeamID int64
	for _, p := range game.Participants {
		if p.RiotID == riotID {
			userTeamID = p.TeamID
			vd.userChampionID, vd.userChampionName = resolve(p.ChampionID)
			vd.found = true
			break
		}
	}
	if !vd.found {
		return vd
	}

	for _, p := range game.Participants {
		if p.TeamID == userTeamID {
			continue
		}
		textID, champName := resolve(p.ChampionID)
		vd.parts = append(vd.parts, components.OpponentView{
			RiotID:       p.RiotID,
			ChampionID:   textID,
			ChampionName: champName,
			PUUID:        p.PUUID,
		})
	}
	return vd
}

const (
	enrichMatchCount = 5
	enrichTimeout    = 10 * time.Second
	enrichParallel   = 5
)

// enrichOpponents fetches recent match data for each opponent and attaches enrichment stats.
// Errors are logged but not propagated (graceful degradation).
func (h *LiveGameHandler) enrichOpponents(ctx context.Context, opponents []components.OpponentView) {
	enrichCtx, cancel := context.WithTimeout(ctx, enrichTimeout)
	defer cancel()

	var wg sync.WaitGroup
	sem := make(chan struct{}, enrichParallel)

	for i := range opponents {
		if opponents[i].PUUID == "" {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			enrichment := h.computeEnrichment(enrichCtx, opponents[idx].PUUID, opponents[idx].ChampionName)
			opponents[idx].Enrichment = &enrichment
		}(i)
	}
	wg.Wait()
}

// computeEnrichment computes enrichment stats for a single opponent from their recent matches.
func (h *LiveGameHandler) computeEnrichment(ctx context.Context, puuid, currentChampName string) models.OpponentEnrichment {
	var e models.OpponentEnrichment

	ids, err := h.Client.FetchMatchIDs(ctx, puuid, h.Config.RiotRegion, h.Config.RiotAPIKey, enrichMatchCount)
	if err != nil {
		h.Logger.Debug("enrichment: failed to fetch match IDs", "puuid", puuid, "error", err)
		return e
	}

	positionCounts := make(map[string]int)
	streakCounting := true
	var lastWin *bool

	for _, matchID := range ids {
		match, fetchErr := h.Client.FetchMatch(ctx, matchID, h.Config.RiotRegion, h.Config.RiotAPIKey)
		if fetchErr != nil {
			continue
		}
		for _, p := range match.Info.Participants {
			if p.PUUID != puuid {
				continue
			}

			// Champion-specific stats
			if p.ChampionName == currentChampName {
				e.ChampionGames++
				if p.Win {
					e.ChampionWins++
				} else {
					e.ChampionLosses++
				}
			}

			// Streak calculation (matches are ordered most-recent first)
			if streakCounting {
				if lastWin == nil {
					w := p.Win
					lastWin = &w
					if p.Win {
						e.WinStreak = 1
					} else {
						e.LossStreak = 1
					}
				} else if *lastWin == p.Win {
					if p.Win {
						e.WinStreak++
					} else {
						e.LossStreak++
					}
				} else {
					streakCounting = false
				}
			}

			// Position frequency for off-role detection
			pos := p.IndividualPosition
			if pos != "" && pos != "Invalid" {
				positionCounts[pos]++
			}
			break
		}
	}

	// Determine most played position
	maxCount := 0
	for pos, count := range positionCounts {
		if count > maxCount {
			maxCount = count
			e.MostPlayedPosition = pos
		}
	}

	// Off-role detection: if they have enough data and never played this
	// position recently, they may be off-role. We approximate by checking
	// if the champion was never seen in their recent matches — meaning
	// they might be on an unfamiliar role. A more accurate check would
	// compare spectator-assigned position vs most-played, but spectator-v5
	// doesn't provide position assignments.
	// Instead: if they have a clear main role (>= 3 of 5 games) and
	// zero games on the current champion, flag as possibly off-role.
	if e.ChampionGames == 0 && len(ids) >= enrichMatchCount && maxCount >= 3 {
		e.PossiblyOffRole = true
	}

	return e
}
