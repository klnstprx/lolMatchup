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

// Bonus constants for fuzzy matching: applied to Levenshtein distance to favor
// names that start with or contain the typed text.
const (
	prefixBonus    = 2 // bonus for a champion whose name starts with typed text
	substringBonus = 1 // bonus if typed text appears anywhere inside name
)

// fuzzyScore computes a weighted Levenshtein distance for a candidate name against
// typed input, applying prefix and substring bonuses. Returns the weighted distance
// and whether the candidate is within threshold.
func fuzzyScore(typed, processedName string, threshold int) (weighted int, ok bool) {
	dist := levenshteinDistance(typed, processedName)
	if dist > threshold+prefixBonus {
		return 0, false
	}
	weighted = dist
	if len(processedName) >= len(typed) && processedName[:len(typed)] == typed {
		weighted -= prefixBonus
	} else if strings.Contains(processedName, typed) {
		weighted -= substringBonus
	}
	if weighted < 0 {
		weighted = 0
	}
	if weighted > threshold {
		return 0, false
	}
	return weighted, true
}

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
	c.ChampionKeyMap = make(map[string]string)
}

// GetPatch returns the current cached patch version.
func (c *Cache) GetPatch() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Patch
}

// SetPatch sets the cached patch version.
func (c *Cache) SetPatch(patch string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Patch = patch
}

// GetChampionMapLen returns the number of entries in the champion name map.
func (c *Cache) GetChampionMapLen() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.ChampionMap)
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
// champion's name starts with (prefix) or contains the user's input (substring).
// Ties in the same weighted distance are broken alphabetically by champion name.
func (c *Cache) SearchChampionName(input string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	typed := preprocessString(input)
	if typed == "" {
		return "", fmt.Errorf("no champion found matching '%s'", input)
	}

	type candidate struct {
		name       string
		championID string
		weighted   int
	}

	var results []candidate

	for name, champID := range c.ChampionMap {
		processedName := preprocessString(name)
		weighted, ok := fuzzyScore(typed, processedName, c.LevenshteinThreshold)
		if !ok {
			continue
		}
		results = append(results, candidate{
			name:       name,
			championID: champID,
			weighted:   weighted,
		})
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no champion found matching '%s'", input)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].weighted != results[j].weighted {
			return results[i].weighted < results[j].weighted
		}
		return results[i].name < results[j].name
	})

	return results[0].championID, nil
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
	c.Champions[champion.Key] = champion
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
	var fuzzy []candidate
	for name := range c.ChampionMap {
		processed := preprocessString(name)
		weighted, ok := fuzzyScore(typed, processed, c.LevenshteinThreshold)
		if !ok {
			continue
		}
		fuzzy = append(fuzzy, candidate{name: name, weighted: weighted})
	}
	if len(fuzzy) == 0 {
		return nil
	}
	sort.Slice(fuzzy, func(i, j int) bool {
		if fuzzy[i].weighted != fuzzy[j].weighted {
			return fuzzy[i].weighted < fuzzy[j].weighted
		}
		// break ties alphabetically
		return fuzzy[i].name < fuzzy[j].name
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

// AutocompleteResult holds enriched data for one autocomplete suggestion.
type AutocompleteResult struct {
	Name      string
	Key       string
	Positions []string
	Roles     []string
}

// AutocompleteRich returns up to 'limit' enriched champion suggestions that best
// match the input, including champion key, positions, and roles.
func (c *Cache) AutocompleteRich(input string, limit int) []AutocompleteResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	typed := preprocessString(input)
	if typed == "" {
		return nil
	}

	var names []string

	// 1) prefix matches
	for name := range c.ChampionMap {
		if strings.HasPrefix(preprocessString(name), typed) {
			names = append(names, name)
		}
	}
	if len(names) > 0 {
		sort.Strings(names)
		if limit > 0 && len(names) > limit {
			names = names[:limit]
		}
		return c.enrichNames(names)
	}

	// 2) substring matches
	for name := range c.ChampionMap {
		if strings.Contains(preprocessString(name), typed) {
			names = append(names, name)
		}
	}
	if len(names) > 0 {
		sort.Strings(names)
		if limit > 0 && len(names) > limit {
			names = names[:limit]
		}
		return c.enrichNames(names)
	}

	// 3) fuzzy fallback
	type candidate struct {
		name     string
		weighted int
	}
	var fuzzy []candidate
	for name := range c.ChampionMap {
		processed := preprocessString(name)
		weighted, ok := fuzzyScore(typed, processed, c.LevenshteinThreshold)
		if !ok {
			continue
		}
		fuzzy = append(fuzzy, candidate{name: name, weighted: weighted})
	}
	if len(fuzzy) == 0 {
		return nil
	}
	sort.Slice(fuzzy, func(i, j int) bool {
		if fuzzy[i].weighted != fuzzy[j].weighted {
			return fuzzy[i].weighted < fuzzy[j].weighted
		}
		return fuzzy[i].name < fuzzy[j].name
	})
	for _, cand := range fuzzy {
		names = append(names, cand.name)
	}
	if limit > 0 && len(names) > limit {
		names = names[:limit]
	}
	return c.enrichNames(names)
}

// enrichNames converts a list of champion names into AutocompleteResults
// by looking up champion data from the cache. Must be called with c.mu held.
func (c *Cache) enrichNames(names []string) []AutocompleteResult {
	results := make([]AutocompleteResult, 0, len(names))
	for _, name := range names {
		key := c.ChampionMap[name]
		r := AutocompleteResult{Name: name, Key: key}
		if champ, ok := c.Champions[key]; ok {
			r.Positions = champ.Positions
			r.Roles = champ.Roles
		}
		results = append(results, r)
	}
	return results
}
