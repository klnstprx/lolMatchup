package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
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

// PlayerPageGET renders the player search page. If a riotID query param is present,
// it performs the lookup server-side and renders the page with results pre-populated.
func (h *PlayerHandler) PlayerPageGET(c *gin.Context) {
	ctx := c.Request.Context()
	prefill := c.Query("riotID")

	var result *components.PlayerResult
	if prefill != "" {
		result = h.lookupPlayer(ctx, prefill)
	}

	cmp := components.PlayerPage(prefill, result)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}

// PlayerGET handles /player requests, returning player info via templ component.
func (h *PlayerHandler) PlayerGET(c *gin.Context) {
	ctx := c.Request.Context()
	riotID, ok := c.GetQuery("riotID")
	if !ok || riotID == "" {
		renderError(c, http.StatusBadRequest, "Summoner identifier is required (nickname#tag).")
		h.Logger.Error("riotID query param missing", "url", c.Request.URL.String())
		return
	}
	parts := strings.SplitN(riotID, "#", 2)
	if len(parts) != 2 {
		renderError(c, http.StatusBadRequest, "Invalid format for Summoner; use nickname#tag.")
		h.Logger.Error("invalid riotID format", "riotID", riotID)
		return
	}
	gameName, tagLine := parts[0], parts[1]

	// Step 1: Resolve Riot ID to PUUID via account-v1
	acct, err := h.Client.FetchAccountByRiotID(ctx, gameName, tagLine, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		switch {
		case errors.Is(err, client.ErrAccountNotFound):
			renderError(c, http.StatusNotFound, fmt.Sprintf("Account '%s' not found.", riotID))
			h.Logger.Debug("account not found", "riotID", riotID)
			return
		case errors.Is(err, client.ErrPermissionDenied):
			renderError(c, http.StatusForbidden, "Permission denied: check your Riot API key and region.")
			h.Logger.Error("permission denied fetching account info", "error", err)
			return
		default:
			h.Logger.Error("error fetching account info", "error", err)
			renderError(c, http.StatusInternalServerError, "Error fetching account data.")
			return
		}
	}

	// Step 2: Fetch summoner data by PUUID via summoner-v4
	player, err := h.Client.FetchSummonerByPUUID(ctx, acct.PUUID, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		if errors.Is(err, client.ErrSummonerNotFound) {
			renderError(c, http.StatusNotFound, fmt.Sprintf("Summoner '%s' not found.", riotID))
			h.Logger.Debug("summoner not found", "riotID", riotID)
			return
		}
		h.Logger.Error("error fetching summoner info", "error", err)
		renderError(c, http.StatusInternalServerError, "Error fetching summoner data.")
		return
	}
	// Step 3: Fetch match history and compute matchup stats
	matches, fullMatches, loaded, total := h.fetchMatchHistory(ctx, acct.PUUID)
	matchups := computeMatchupStats(fullMatches, acct.PUUID)
	fetchedAt := time.Now()

	cmp := components.PlayerComponent(acct, player, matches, matchups, h.Config, loaded, total, fetchedAt)
	c.Render(http.StatusOK, renderer.New(ctx, http.StatusOK, cmp))
}

// lookupPlayer performs the full player lookup (account → summoner → matches).
// On failure, returns a PlayerResult with the Error field set.
func (h *PlayerHandler) lookupPlayer(ctx context.Context, riotID string) *components.PlayerResult {
	parts := strings.SplitN(riotID, "#", 2)
	if len(parts) != 2 {
		return &components.PlayerResult{Error: "Invalid format for Summoner; use nickname#tag."}
	}
	gameName, tagLine := parts[0], parts[1]

	acct, err := h.Client.FetchAccountByRiotID(ctx, gameName, tagLine, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		h.Logger.Debug("player page lookup: account error", "riotID", riotID, "error", err)
		switch {
		case errors.Is(err, client.ErrAccountNotFound):
			return &components.PlayerResult{Error: fmt.Sprintf("Account '%s' not found.", riotID)}
		case errors.Is(err, client.ErrPermissionDenied):
			return &components.PlayerResult{Error: "Permission denied: check your Riot API key and region."}
		default:
			return &components.PlayerResult{Error: "Error fetching account data."}
		}
	}

	player, err := h.Client.FetchSummonerByPUUID(ctx, acct.PUUID, h.Config.RiotRegion, h.Config.RiotAPIKey)
	if err != nil {
		h.Logger.Debug("player page lookup: summoner error", "riotID", riotID, "error", err)
		if errors.Is(err, client.ErrSummonerNotFound) {
			return &components.PlayerResult{Error: fmt.Sprintf("Summoner '%s' not found.", riotID)}
		}
		return &components.PlayerResult{Error: "Error fetching summoner data."}
	}

	matches, fullMatches, loaded, total := h.fetchMatchHistory(ctx, acct.PUUID)
	matchups := computeMatchupStats(fullMatches, acct.PUUID)

	return &components.PlayerResult{
		Account:       acct,
		Summoner:      player,
		Matches:       matches,
		Matchups:      matchups,
		Config:        h.Config,
		FetchedAt:     time.Now(),
		MatchesLoaded: loaded,
		MatchesTotal:  total,
	}
}

const matchHistoryCount = 10

// fetchMatchHistory retrieves recent matches for a player and extracts summaries.
// Returns condensed summaries, full match DTOs, and counts of loaded/total matches.
// Errors are logged but not surfaced — match history is non-critical.
func (h *PlayerHandler) fetchMatchHistory(ctx context.Context, puuid string) ([]models.MatchSummary, []models.MatchDTO, int, int) {
	ids, err := h.Client.FetchMatchIDs(ctx, puuid, h.Config.RiotRegion, h.Config.RiotAPIKey, matchHistoryCount)
	if err != nil {
		h.Logger.Warn("failed to fetch match IDs", "error", err)
		return nil, nil, 0, 0
	}
	if len(ids) == 0 {
		return nil, nil, 0, 0
	}

	// Fetch all matches concurrently
	type result struct {
		index   int
		summary models.MatchSummary
		match   models.MatchDTO
		ok      bool
	}
	results := make([]result, len(ids))
	var wg sync.WaitGroup

	for i, matchID := range ids {
		wg.Add(1)
		go func(idx int, mid string) {
			defer wg.Done()
			match, err := h.Client.FetchMatch(ctx, mid, h.Config.RiotRegion, h.Config.RiotAPIKey)
			if err != nil {
				h.Logger.Debug("failed to fetch match", "matchId", mid, "error", err)
				return
			}
			// Find the target player's participant data
			for _, p := range match.Info.Participants {
				if p.PUUID == puuid {
					results[idx] = result{
						index: idx,
						ok:    true,
						match: match,
						summary: models.MatchSummary{
							MatchID:       mid,
							ChampionName:  p.ChampionName,
							ChampionID:    p.ChampionID,
							Win:           p.Win,
							Kills:         p.Kills,
							Deaths:        p.Deaths,
							Assists:       p.Assists,
							CS:            p.TotalMinionsKilled + p.NeutralMinionsKilled,
							Position:      p.IndividualPosition,
							GameDuration:  match.Info.GameDuration,
							GameStartTime: match.Info.GameStartTimestamp,
							Damage:        p.TotalDamageDealtToChampions,
							Gold:          p.GoldEarned,
							VisionScore:   p.VisionScore,
							Items:         [7]int{p.Item0, p.Item1, p.Item2, p.Item3, p.Item4, p.Item5, p.Item6},
							QueueID:       match.Info.QueueID,
						},
					}
					break
				}
			}
		}(i, matchID)
	}
	wg.Wait()

	// Collect successful results in order
	var summaries []models.MatchSummary
	var fullMatches []models.MatchDTO
	for _, r := range results {
		if r.ok {
			summaries = append(summaries, r.summary)
			fullMatches = append(fullMatches, r.match)
		}
	}
	return summaries, fullMatches, len(summaries), len(ids)
}

// computeMatchupStats aggregates lane matchup records from full match data.
func computeMatchupStats(matches []models.MatchDTO, puuid string) []models.MatchupRecord {
	type key struct {
		playerChamp string
		enemyChamp  string
	}
	stats := make(map[key]*models.MatchupRecord)

	for _, match := range matches {
		// Find the target player
		var player *models.MatchParticipant
		for i, p := range match.Info.Participants {
			if p.PUUID == puuid {
				player = &match.Info.Participants[i]
				break
			}
		}
		if player == nil || player.IndividualPosition == "" || player.IndividualPosition == "Invalid" {
			continue
		}

		// Find the enemy laner (same position, different team)
		for _, p := range match.Info.Participants {
			if p.TeamID != player.TeamID && p.IndividualPosition == player.IndividualPosition {
				k := key{playerChamp: player.ChampionName, enemyChamp: p.ChampionName}
				if stats[k] == nil {
					stats[k] = &models.MatchupRecord{
						PlayerChampion: player.ChampionName,
						EnemyChampion:  p.ChampionName,
					}
				}
				if player.Win {
					stats[k].Wins++
				} else {
					stats[k].Losses++
				}
				stats[k].Games++
				break
			}
		}
	}

	// Convert map to sorted slice (most games first)
	records := make([]models.MatchupRecord, 0, len(stats))
	for _, r := range stats {
		records = append(records, *r)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Games > records[j].Games
	})
	return records
}
