package client

import (
   "context"
   "encoding/json"
   "fmt"
   "errors"
   "io"
   "net/http"

   "github.com/charmbracelet/log"
   "github.com/klnstprx/lolMatchup/models"
)

type Client struct {
	HTTPClient *http.Client
	Logger     *log.Logger
}

// ErrChampionNotFound indicates the requested champion was not found in Data Dragon.
var ErrChampionNotFound = errors.New("champion not found")

// FetchLatestPatch retrieves the latest game version from the DDragon API.
func (c *Client) FetchLatestPatch(ctx context.Context, ddragonVersionURL string) (string, error) {
	c.Logger.Debug("Fetching latest patch", "url", ddragonVersionURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ddragonVersionURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch versions from DDragon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d fetching versions", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var versions []string
	if err = json.Unmarshal(body, &versions); err != nil {
		return "", fmt.Errorf("failed to parse JSON data: %w", err)
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found in response")
	}

	return versions[0], nil
}

// FetchChampionNameIDMap retrieves a map of champion names -> champion IDs.
func (c *Client) FetchChampionNameIDMap(ctx context.Context, ddragonURL, patchNumber, languageCode string) (map[string]string, error) {
	targetURL := fmt.Sprintf("%s%s/data/%s/champion.json", ddragonURL, patchNumber, languageCode)
	c.Logger.Debug("Fetching champion name/ID map", "url", targetURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch champion list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d fetching champion list", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var championList models.ChampionList
	if err = json.Unmarshal(body, &championList); err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %w", err)
	}

	championMap := make(map[string]string, len(championList.Data))
	for _, champ := range championList.Data {
		championMap[champ.Name] = champ.ID
	}

	return championMap, nil
}

// FetchChampionData fetches detailed champion information for a given champion ID.
func (c *Client) FetchChampionData(ctx context.Context, championID, ddragonURLData string) (models.Champion, error) {
	var champion models.Champion
	targetURL := fmt.Sprintf("%s%s.json", ddragonURLData, championID)
	c.Logger.Debug("Fetching champion data", "url", targetURL, "champID", championID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return champion, fmt.Errorf("failed to create request for champion %s: %w", championID, err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return champion, fmt.Errorf("failed to fetch data for champion %s: %w", championID, err)
	}
	defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
       if resp.StatusCode == http.StatusNotFound {
           return champion, ErrChampionNotFound
       }
       return champion, fmt.Errorf("unexpected status code %d fetching champion data", resp.StatusCode)
   }

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return champion, fmt.Errorf("failed to read response body: %w", err)
	}

   var root models.Root
   if err = json.Unmarshal(body, &root); err != nil {
       return champion, fmt.Errorf("failed to parse JSON data: %w", err)
   }
   // If no data present, the champion was not found
   if len(root.Data) == 0 {
       return champion, ErrChampionNotFound
   }
   // There's usually just one champion in the "Data" map; retrieve first.
   for _, champ := range root.Data {
       champion = champ
       break
   }
   return champion, nil
}
