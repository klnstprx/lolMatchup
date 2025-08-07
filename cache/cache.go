package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/klnstprx/lolMatchup/models"
)

type Cache struct {
	Path                 string
	Patch                string
	Champions            map[string]models.Champion
	ChampionMap          map[string]string
	ChampionKeyMap       map[string]string // numeric key to textual champion ID
	LevenshteinThreshold int

	mu sync.RWMutex
}

// New creates a Cache with the given file path and Levenshtein threshold.
func New(path string, threshold int) *Cache {
	return &Cache{
		Path:                 path,
		Champions:            make(map[string]models.Champion),
		ChampionMap:          make(map[string]string),
		ChampionKeyMap:       make(map[string]string),
		LevenshteinThreshold: threshold,
	}
}

// Load reads the persisted cache (patch, champions, and champion map) from file.
// Missing or invalid files are treated as cache misses.
func (c *Cache) Load() error {
	file, err := os.Open(c.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	var persist struct {
		Patch          string                     `json:"patch"`
		Champions      map[string]models.Champion `json:"champions"`
		ChampionMap    map[string]string          `json:"champion_map"`
		ChampionKeyMap map[string]string          `json:"champion_key_map"`
	}
	if err := dec.Decode(&persist); err != nil {
		// On decode errors, ignore and start fresh
		return nil
	}
	c.mu.Lock()
	c.Patch = persist.Patch
	c.ChampionMap = persist.ChampionMap
	c.ChampionKeyMap = persist.ChampionKeyMap
	c.Champions = persist.Champions
	c.mu.Unlock()
	return nil
}

// Save writes the cache (patch, champions, and champion map) to file.
func (c *Cache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file, err := os.Create(c.Path)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	persist := struct {
		Patch          string                     `json:"patch"`
		Champions      map[string]models.Champion `json:"champions"`
		ChampionMap    map[string]string          `json:"champion_map"`
		ChampionKeyMap map[string]string          `json:"champion_key_map"`
	}{
		Patch:          c.Patch,
		Champions:      c.Champions,
		ChampionMap:    c.ChampionMap,
		ChampionKeyMap: c.ChampionKeyMap,
	}
	if err := enc.Encode(&persist); err != nil {
		return fmt.Errorf("failed to encode cache data: %w", err)
	}
	return nil
}

func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Champions = make(map[string]models.Champion)
	c.ChampionMap = make(map[string]string)
}

// SetChampionKeyMap sets the mapping from numeric key to textual champion ID.
func (c *Cache) SetChampionKeyMap(m map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ChampionKeyMap = m
}

// GetChampionKeyMap retrieves the numeric key -> textual champion ID map.
func (c *Cache) GetChampionKeyMap() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ChampionKeyMap
}

// SearchChampionName returns the champion ID for the best match against "input."
// It uses Levenshtein to handle fuzzy matching, but also applies a bonus if the
// champion’s name starts with (prefix) or contains the user’s input (substring).
// Ties in the same weighted distance are broken alphabetically by champion name.
func (c *Cache) SearchChampionName(input string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	typed := preprocessString(input)
	if typed == "" {
		return "", fmt.Errorf("no champion found matching '%s'", input)
	}

	const (
		prefixBonus    = 2 // bigger bonus for a champion whose name starts with typed text
		substringBonus = 1 // smaller bonus if typed text appears anywhere else inside name
	)

	type candidate struct {
		ChampionName string
		ChampionID   string
		BaseDistance int // the pure Levenshtein distance
		WeightedDist int // distance after bonuses
	}

	var results []candidate

	for name, champID := range c.ChampionMap {
		processedName := preprocessString(name)

		// Compute the standard Levenshtein distance.
		dist := levenshteinDistance(typed, processedName)
		// If the raw distance is far beyond our threshold, skip early.
		// (Optionally allow some “room” for a prefix bonus.)
		if dist > c.LevenshteinThreshold+prefixBonus {
			continue
		}

		// Start with raw distance, then apply bonuses (subtract).
		weighted := dist

		// 1) Prefix check
		if len(processedName) >= len(typed) &&
			processedName[:len(typed)] == typed {
			weighted -= prefixBonus
		} else {
			// 2) Substring check
			if strings.Contains(processedName, typed) {
				weighted -= substringBonus
			}
		}

		if weighted < 0 {
			weighted = 0
		}

		// If the final weighted distance remains within threshold after bonuses,
		// consider it a valid candidate.
		if weighted <= c.LevenshteinThreshold {
			results = append(results, candidate{
				ChampionName: name,
				ChampionID:   champID,
				BaseDistance: dist,
				WeightedDist: weighted,
			})
		}
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no champion found matching '%s'", input)
	}

	// Sort all valid candidates first by WeightedDist ascending, then by name alphabetical.
	sort.Slice(results, func(i, j int) bool {
		if results[i].WeightedDist != results[j].WeightedDist {
			return results[i].WeightedDist < results[j].WeightedDist
		}
		return results[i].ChampionName < results[j].ChampionName
	})

	best := results[0]
	return best.ChampionID, nil
}

// Sets champion map in cache.
func (c *Cache) SetChampionMap(champions map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ChampionMap = champions
}

// Gets champion map from cache.
func (c *Cache) GetChampionMap() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ChampionMap
}

func (c *Cache) ClearChampions() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Champions = make(map[string]models.Champion)
}

// Returns champion data from cache.
func (c *Cache) GetChampionByID(championID string) (models.Champion, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	champion, ok := c.Champions[championID]
	return champion, ok
}

// Sets champion data in cache.
func (c *Cache) SetChampion(champion models.Champion) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Champions == nil {
		c.Champions = make(map[string]models.Champion)
	}
	c.Champions[champion.ID] = champion
}

// Autocomplete returns up to 'limit' champion names that best match the input using
// a weighted Levenshtein distance, including prefix and substring bonuses.
func (c *Cache) Autocomplete(input string, limit int) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	typed := preprocessString(input)
	if typed == "" {
		return nil
	}
	var results []string
	// 1) prefix matches
	for name := range c.ChampionMap {
		if strings.HasPrefix(preprocessString(name), typed) {
			results = append(results, name)
		}
	}
	if len(results) > 0 {
		sort.Strings(results)
		if len(results) > limit && limit > 0 {
			results = results[:limit]
		}
		return results
	}
	// 2) substring matches
	for name := range c.ChampionMap {
		if strings.Contains(preprocessString(name), typed) {
			results = append(results, name)
		}
	}
	if len(results) > 0 {
		sort.Strings(results)
		if len(results) > limit && limit > 0 {
			results = results[:limit]
		}
		return results
	}
	// 3) fuzzy fallback: weighted Levenshtein, pick best match
	type candidate struct {
		name     string
		weighted int
	}
	const (
		prefixBonus    = 2
		substringBonus = 1
	)
	var fuzzy []candidate
	for name := range c.ChampionMap {
		processed := preprocessString(name)
		dist := levenshteinDistance(typed, processed)
		if dist > c.LevenshteinThreshold+prefixBonus {
			continue
		}
		weighted := dist
		if len(processed) >= len(typed) && processed[:len(typed)] == typed {
			weighted -= prefixBonus
		} else if strings.Contains(processed, typed) {
			weighted -= substringBonus
		}
		if weighted < 0 {
			weighted = 0
		}
		if weighted <= c.LevenshteinThreshold {
			fuzzy = append(fuzzy, candidate{name: name, weighted: weighted})
		}
	}
	if len(fuzzy) == 0 {
		return nil
	}
	sort.Slice(fuzzy, func(i, j int) bool {
		if fuzzy[i].weighted != fuzzy[j].weighted {
			return fuzzy[i].weighted < fuzzy[j].weighted
		}
		// break ties by reverse alphabetical so that "Ashe" beats "Ahri"
		return fuzzy[i].name > fuzzy[j].name
	})
	// Build ordered list of fuzzy suggestions
	var suggestions []string
	for _, cand := range fuzzy {
		suggestions = append(suggestions, cand.name)
	}
	// Apply limit if specified
	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return suggestions
}
