package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/klnstprx/lolMatchup/models"
)

// Client is a unified client for all API interactions.
type Client struct {
	HTTPClient        *http.Client
	Logger            *log.Logger
	ChampionDataURL   string
	DDragonVersionURL string
	RiotAPIBaseURL    string // when non-empty, overrides Riot API hostname for mock/dev use
}

// riotURL builds the base URL for Riot API calls. When RiotAPIBaseURL is set,
// it is used directly (ignoring hostPrefix). Otherwise, the standard
// https://{hostPrefix}.api.riotgames.com format is used.
func (c *Client) riotURL(hostPrefix string) string {
	if c.RiotAPIBaseURL != "" {
		return strings.TrimRight(c.RiotAPIBaseURL, "/")
	}
	return fmt.Sprintf("https://%s.api.riotgames.com", hostPrefix)
}

// APIError represents a non-200 HTTP response from an API call.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Body)
}

// doJSON performs a GET request, reads the response, and unmarshals JSON into target.
// For non-200 responses it returns an *APIError. If riotAPIKey is non-empty, the
// X-Riot-Token header is set.
func (c *Client) doJSON(ctx context.Context, url, riotAPIKey string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if riotAPIKey != "" {
		req.Header.Set("X-Riot-Token", riotAPIKey)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	return nil
}

// mapAPIError maps an *APIError to domain-specific sentinel errors.
// Returns the original error unchanged if it is not an *APIError.
func mapAPIError(err error, notFoundErr error) error {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return err
	}
	switch apiErr.StatusCode {
	case http.StatusNotFound:
		return notFoundErr
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrPermissionDenied, apiErr.Body)
	default:
		return err
	}
}

// Sentinel errors.
var (
	ErrSummonerNotFound = errors.New("summoner not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrGameNotFound     = errors.New("game not found")
	ErrAccountNotFound  = errors.New("account not found")
	ErrChampionNotFound = errors.New("champion not found")
	ErrMatchNotFound    = errors.New("match not found")
)

// RegionToCluster maps Riot API regional routing values to their continental cluster.
var RegionToCluster = map[string]string{
	"na1": "americas", "br1": "americas", "la1": "americas", "la2": "americas",
	"euw1": "europe", "eun1": "europe", "ru": "europe", "tr1": "europe",
	"kr": "asia", "jp1": "asia",
	"oc1": "sea", "sg2": "sea", "tw2": "sea", "vn2": "sea",
}

// SummonerDTO represents the data returned by the Summoner API.
type SummonerDTO struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	PUUID         string `json:"puuid"`
	Name          string `json:"name"`
	ProfileIconID int    `json:"profileIconId"`
	RevisionDate  int64  `json:"revisionDate"`
	SummonerLevel int64  `json:"summonerLevel"`
}

// FetchSummonerByPUUID retrieves summoner information by encrypted PUUID.
func (c *Client) FetchSummonerByPUUID(ctx context.Context, puuid, riotRegion, riotAPIKey string) (SummonerDTO, error) {
	var summoner SummonerDTO
	reqURL := fmt.Sprintf("%s/lol/summoner/v4/summoners/by-puuid/%s", c.riotURL(riotRegion), url.PathEscape(puuid))
	if err := c.doJSON(ctx, reqURL, riotAPIKey, &summoner); err != nil {
		return summoner, mapAPIError(err, ErrSummonerNotFound)
	}
	return summoner, nil
}

// AccountDTO holds encrypted PUUID and associated Riot account info.
type AccountDTO struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// FetchAccountByRiotID retrieves account information (incl. puuid) via gameName/tagLine.
func (c *Client) FetchAccountByRiotID(ctx context.Context, gameName, tagLine, riotRegion, riotAPIKey string) (AccountDTO, error) {
	var acct AccountDTO
	// Determine the regional cluster for account-v1.
	cluster, ok := RegionToCluster[riotRegion]
	if !ok {
		cluster = riotRegion
	}
	reqURL := fmt.Sprintf(
		"%s/riot/account/v1/accounts/by-riot-id/%s/%s",
		c.riotURL(cluster), url.PathEscape(gameName), url.PathEscape(tagLine),
	)
	if err := c.doJSON(ctx, reqURL, riotAPIKey, &acct); err != nil {
		return acct, mapAPIError(err, ErrAccountNotFound)
	}
	return acct, nil
}

// FetchCurrentGameByPUUID retrieves current game info using encrypted PUUID (spectator v5).
func (c *Client) FetchCurrentGameByPUUID(ctx context.Context, puuid, riotRegion, riotAPIKey string) (models.CurrentGameInfo, error) {
	var game models.CurrentGameInfo
	reqURL := fmt.Sprintf("%s/lol/spectator/v5/active-games/by-summoner/%s", c.riotURL(riotRegion), url.PathEscape(puuid))
	if err := c.doJSON(ctx, reqURL, riotAPIKey, &game); err != nil {
		return game, mapAPIError(err, ErrGameNotFound)
	}
	return game, nil
}

// FetchMatchIDs retrieves recent match IDs for a player by PUUID (match-v5, cluster routing).
func (c *Client) FetchMatchIDs(ctx context.Context, puuid, riotRegion, riotAPIKey string, count int) ([]string, error) {
	cluster, ok := RegionToCluster[riotRegion]
	if !ok {
		cluster = riotRegion
	}
	var ids []string
	reqURL := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?count=%d", c.riotURL(cluster), url.PathEscape(puuid), count)
	if err := c.doJSON(ctx, reqURL, riotAPIKey, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// FetchMatch retrieves full match data by match ID (match-v5, cluster routing).
func (c *Client) FetchMatch(ctx context.Context, matchID, riotRegion, riotAPIKey string) (models.MatchDTO, error) {
	cluster, ok := RegionToCluster[riotRegion]
	if !ok {
		cluster = riotRegion
	}
	var match models.MatchDTO
	reqURL := fmt.Sprintf("%s/lol/match/v5/matches/%s", c.riotURL(cluster), url.PathEscape(matchID))
	if err := c.doJSON(ctx, reqURL, riotAPIKey, &match); err != nil {
		return match, mapAPIError(err, ErrMatchNotFound)
	}
	return match, nil
}

// FetchChampionData fetches detailed champion information for a given champion ID.
func (c *Client) FetchChampionData(ctx context.Context, championID string) (models.Champion, error) {
	var champion models.Champion
	reqURL := fmt.Sprintf("%schampions/%s.json", c.ChampionDataURL, url.PathEscape(championID))
	c.Logger.Debug("Fetching champion data", "url", reqURL, "champID", championID)
	if err := c.doJSON(ctx, reqURL, "", &champion); err != nil {
		return champion, mapAPIError(err, ErrChampionNotFound)
	}
	return champion, nil
}

// FetchChampionList fetches a map of all champions.
func (c *Client) FetchChampionList(ctx context.Context) (map[string]models.Champion, error) {
	var champions map[string]models.Champion
	reqURL := fmt.Sprintf("%schampions.json", c.ChampionDataURL)
	c.Logger.Debug("Fetching champion list", "url", reqURL)
	if err := c.doJSON(ctx, reqURL, "", &champions); err != nil {
		return champions, err
	}
	return champions, nil
}

// FetchLatestPatch retrieves the latest game version from the DDragon API.
func (c *Client) FetchLatestPatch(ctx context.Context) (string, error) {
	c.Logger.Debug("Fetching latest patch", "url", c.DDragonVersionURL)
	var versions []string
	if err := c.doJSON(ctx, c.DDragonVersionURL, "", &versions); err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found in response")
	}
	return versions[0], nil
}
