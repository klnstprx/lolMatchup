package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/klnstprx/lolMatchup/models"
)

type Client struct {
	HTTPClient *http.Client
	Logger     *log.Logger
}

// FetchLatestPatch fetches the latest game version from the DDragon API.
func (c *Client) FetchLatestPatch(ctx context.Context, ddragonVersionURL string) (string, error) {
	c.Logger.Debug("Querying...", "url", ddragonVersionURL)

	resp, err := c.HTTPClient.Get(ddragonVersionURL)
	if err != nil {
		return "", fmt.Errorf("error fetching versions from DDragon API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d fetching versions", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var versions []string
	err = json.Unmarshal(body, &versions)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON data: %v", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found in response")
	}

	latestVersion := versions[0]
	return latestVersion, nil
}

// FetchChampionNameIDMap fetches a map of champion names to their IDs.
func (c *Client) FetchChampionNameIDMap(ctx context.Context, ddragonURL, patchNumber, languageCode string) (map[string]string, error) {
	targetURL := fmt.Sprintf("%s%s/data/%s/champion.json", ddragonURL, patchNumber, languageCode)
	c.Logger.Debug("Querying...", "url", targetURL)

	resp, err := c.HTTPClient.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching champion list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d fetching champion list", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var championList models.ChampionList
	err = json.Unmarshal(body, &championList)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON data: %v", err)
	}

	championMap := make(map[string]string)
	for _, champ := range championList.Data {
		championMap[champ.Name] = champ.ID
	}

	return championMap, nil
}

// FetchChampionData fetches champion data from the DDragon API.
func (c *Client) FetchChampionData(ctx context.Context, championID, ddragonURLData string) (models.Champion, error) {
	var champion models.Champion
	targetURL := fmt.Sprintf("%s%s.json", ddragonURLData, championID)
	c.Logger.Debug("Querying...", "url", targetURL)

	resp, err := c.HTTPClient.Get(targetURL)
	if err != nil {
		return champion, fmt.Errorf("error fetching data for champion %s: %v", championID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return champion, fmt.Errorf("champion '%s' not found", championID)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return champion, fmt.Errorf("error reading response body: %v", err)
	}

	var root models.Root
	err = json.Unmarshal(body, &root)
	if err != nil {
		return champion, fmt.Errorf("error parsing JSON data: %v", err)
	}

	for _, champ := range root.Data {
		champion = champ
		break
	}
	return champion, nil
}
