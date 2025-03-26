package cache

import (
	"encoding/gob"
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
	LevenshteinThreshold int

	mu sync.RWMutex
}

func New(path string, threshold int) *Cache {
	return &Cache{
		Path:                 path,
		Champions:            make(map[string]models.Champion),
		ChampionMap:          make(map[string]string),
		LevenshteinThreshold: threshold,
	}
}

func (c *Cache) Load() error {
	file, err := os.Open(c.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cache file not found: %w", err)
		}
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	err = dec.Decode(c)
	if err != nil {
		return fmt.Errorf("failed to decode cache file: %w", err)
	}
	return nil
}

func (c *Cache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file, err := os.Create(c.Path)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(c)
	if err != nil {
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
