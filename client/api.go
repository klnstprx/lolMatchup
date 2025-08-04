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
	ErrPermissionDenied = errors.New("")
)

// FetchSummonerByName retrieves summoner information by summoner name.
func (c *Client) FetchSummonerByName(ctx context.Context, summonerName, riotRegion, riotAPIKey string) (SummonerDTO, error) {
	var summoner SummonerDTO
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/summoner/v4/summoners/by-name/%s?api_key=%s", riotRegion, summonerName, riotAPIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return summoner, fmt.Errorf("failed to create summoner request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return summoner, fmt.Errorf("failed to fetch summoner data: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return summoner, ErrSummonerNotFound
		}
		return summoner, fmt.Errorf("unexpected status code %d fetching summoner data", resp.StatusCode)
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

// FetchActiveGame retrieves the active game for a given encryptedSummonerID.
func (c *Client) FetchActiveGame(ctx context.Context, encryptedSummonerID, riotRegion, riotAPIKey string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://%s.api.riotgames.com/lol/spectator/v4/active-games/by-summoner/%s?api_key=%s", riotRegion, encryptedSummonerID, riotAPIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create active game request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active game: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("unexpected status code %d fetching active game", resp.StatusCode)
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
