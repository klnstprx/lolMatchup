package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// ErrSummonerNotFound indicates the requested summoner was not found.
var (
	ErrSummonerNotFound = errors.New("summoner not found")
	// ErrPermissionDenied indicates the Riot API key or region is invalid or forbidden.
	ErrPermissionDenied = errors.New("permission denied")
)

// FetchSummonerByName retrieves summoner information by summoner name.
func (c *Client) FetchSummonerByName(ctx context.Context, summonerName, riotRegion, riotAPIKey string) (SummonerDTO, error) {
	var summoner SummonerDTO
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/summoner/v4/summoners/by-name/%s", riotRegion, summonerName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return summoner, fmt.Errorf("failed to create summoner request: %w", err)
	}
	// Use header authentication for Riot API key
	req.Header.Set("X-Riot-Token", riotAPIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return summoner, fmt.Errorf("failed to fetch summoner data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// capture response body for debugging
		data, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case http.StatusNotFound:
			return summoner, ErrSummonerNotFound
		case http.StatusForbidden:
			return summoner, fmt.Errorf("%w: %s", ErrPermissionDenied, string(data))
		}
		return summoner, fmt.Errorf("unexpected status code %d fetching summoner data: %s", resp.StatusCode, string(data))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return summoner, fmt.Errorf("failed to read summoner response: %w", err)
	}
	if err := json.Unmarshal(body, &summoner); err != nil {
		return summoner, fmt.Errorf("failed to parse summoner JSON: %w", err)
	}
	return summoner, nil
}

// ErrGameNotFound indicates the summoner is not currently in a game.
var ErrGameNotFound = errors.New("game not found")

// AccountDTO holds encrypted PUUID and associated Riot account info.
type AccountDTO struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// ErrAccountNotFound indicates the account (riot id) was not found.
var ErrAccountNotFound = errors.New("account not found")

// FetchActiveGame retrieves the active game for a given encryptedSummonerID.
func (c *Client) FetchActiveGame(ctx context.Context, encryptedSummonerID, riotRegion, riotAPIKey string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/spectator/v4/active-games/by-summoner/%s", riotRegion, encryptedSummonerID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create active game request: %w", err)
	}
	// Use header authentication for Riot API key
	req.Header.Set("X-Riot-Token", riotAPIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active game: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// capture response body for debugging
		data, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, ErrGameNotFound
		case http.StatusForbidden:
			return nil, fmt.Errorf("%w: %s", ErrPermissionDenied, string(data))
		}
		return nil, fmt.Errorf("unexpected status code %d fetching active game: %s", resp.StatusCode, string(data))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read active game response: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse active game JSON: %w", err)
	}
	return result, nil
}

// FetchAccountByRiotID retrieves account information (incl. puuid) via gameName/tagLine.
func (c *Client) FetchAccountByRiotID(ctx context.Context, gameName, tagLine, riotRegion, riotAPIKey string) (AccountDTO, error) {
	var acct AccountDTO
	// account-v1 is global but served on platform clusters by region groups (Americas, Europe, Asia, SEA)
	// Determine the platform cluster for account-v1: use region group names lowercase
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
		// fallback to the raw region if it doesn't map
		cluster = riotRegion
	}
	url := fmt.Sprintf(
		"https://%s.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s",
		cluster, gameName, tagLine,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return acct, fmt.Errorf("failed to create account request: %w", err)
	}
	req.Header.Set("X-Riot-Token", riotAPIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return acct, fmt.Errorf("failed to fetch account data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// capture response body for debugging
		data, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case http.StatusNotFound:
			return acct, ErrAccountNotFound
		case http.StatusForbidden:
			return acct, fmt.Errorf("%w: %s", ErrPermissionDenied, string(data))
		}
		return acct, fmt.Errorf("unexpected status code %d fetching account data: %s", resp.StatusCode, string(data))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return acct, fmt.Errorf("failed to read account response: %w", err)
	}
	if err := json.Unmarshal(body, &acct); err != nil {
		return acct, fmt.Errorf("failed to parse account JSON: %w", err)
	}
	return acct, nil
}

// FetchCurrentGameByPUUID retrieves current game info using encrypted PUUID (spectator v5).
func (c *Client) FetchCurrentGameByPUUID(ctx context.Context, puuid, riotRegion, riotAPIKey string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/spectator/v5/active-games/by-summoner/%s", riotRegion, puuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create spectator v5 request: %w", err)
	}
	req.Header.Set("X-Riot-Token", riotAPIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current game by puuid: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// capture response body for debugging
		data, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, ErrGameNotFound
		case http.StatusForbidden:
			return nil, fmt.Errorf("%w: %s", ErrPermissionDenied, string(data))
		}
		return nil, fmt.Errorf("unexpected status code %d fetching spectator v5 data: %s", resp.StatusCode, string(data))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read spectator v5 response: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse spectator v5 JSON: %w", err)
	}
	return result, nil
}
