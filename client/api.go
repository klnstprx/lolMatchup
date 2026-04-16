package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/klnstprx/lolMatchup/models"
)

// Client is a unified client for all API interactions.
type Client struct {
	HTTPClient        *http.Client
	Logger            *log.Logger
	ChampionDataURL   string
	DDragonVersionURL string
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
)

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

// FetchSummonerByName retrieves summoner information by summoner name.
func (c *Client) FetchSummonerByName(ctx context.Context, summonerName, riotRegion, riotAPIKey string) (SummonerDTO, error) {
	var summoner SummonerDTO
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/summoner/v4/summoners/by-name/%s", riotRegion, summonerName)
	if err := c.doJSON(ctx, url, riotAPIKey, &summoner); err != nil {
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
	var cluster string
	switch riotRegion {
	case "na1", "br1", "la1", "la2":
		cluster = "americas"
	case "euw1", "eun1", "ru", "tr1":
		cluster = "europe"
	case "kr", "jp1":
		cluster = "asia"
	case "oc1", "sg2", "tw2", "vn2":
		cluster = "sea"
	default:
		cluster = riotRegion
	}
	url := fmt.Sprintf(
		"https://%s.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s",
		cluster, gameName, tagLine,
	)
	if err := c.doJSON(ctx, url, riotAPIKey, &acct); err != nil {
		return acct, mapAPIError(err, ErrAccountNotFound)
	}
	return acct, nil
}

// FetchCurrentGameByPUUID retrieves current game info using encrypted PUUID (spectator v5).
func (c *Client) FetchCurrentGameByPUUID(ctx context.Context, puuid, riotRegion, riotAPIKey string) (models.CurrentGameInfo, error) {
	var game models.CurrentGameInfo
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/spectator/v5/active-games/by-summoner/%s", riotRegion, puuid)
	if err := c.doJSON(ctx, url, riotAPIKey, &game); err != nil {
		return game, mapAPIError(err, ErrGameNotFound)
	}
	return game, nil
}

// FetchChampionData fetches detailed champion information for a given champion ID.
func (c *Client) FetchChampionData(ctx context.Context, championID string) (models.Champion, error) {
	var champion models.Champion
	url := fmt.Sprintf("%schampions/%s.json", c.ChampionDataURL, championID)
	c.Logger.Debug("Fetching champion data", "url", url, "champID", championID)
	if err := c.doJSON(ctx, url, "", &champion); err != nil {
		return champion, mapAPIError(err, ErrChampionNotFound)
	}
	return champion, nil
}

// FetchChampionList fetches a map of all champions.
func (c *Client) FetchChampionList(ctx context.Context) (map[string]models.Champion, error) {
	var champions map[string]models.Champion
	url := fmt.Sprintf("%schampions.json", c.ChampionDataURL)
	c.Logger.Debug("Fetching champion list", "url", url)
	if err := c.doJSON(ctx, url, "", &champions); err != nil {
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
