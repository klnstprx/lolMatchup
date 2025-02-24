package cache

import (
	"encoding/gob"
	"fmt"
	"os"
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
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	err = dec.Decode(c)
	if err != nil {
		return fmt.Errorf("error decoding file: %v", err)
	}
	return nil
}

func (c *Cache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file, err := os.Create(c.Path)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(c)
	if err != nil {
		return fmt.Errorf("error encoding file: %v", err)
	}
	return nil
}

func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Champions = make(map[string]models.Champion)
	c.ChampionMap = make(map[string]string)
}

/*
Searches for a champion's name in the saved map of all champions.
Name with lowest Levenshtein distance is returned along with its ID.
c.LevenshteinThreshold is the maximum allowed distance.
*/
func (c *Cache) SearchChampionName(input string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	input = preprocessString(input)
	minDistance := -1
	closestMatch := ""

	for championName := range c.ChampionMap {
		preprocessedChampion := preprocessString(championName)
		distance := levenshteinDistance(input, preprocessedChampion)
		if minDistance == -1 || distance < minDistance {
			minDistance = distance
			closestMatch = championName
		}
		// Break early if exact match
		if distance == 0 {
			break
		}
	}

	if minDistance == -1 {
		return "", fmt.Errorf("no champion found matching '%s'", input)
	}

	if minDistance > c.LevenshteinThreshold {
		return "", fmt.Errorf("threshold (%d) not met for: '%s'", c.LevenshteinThreshold, input)
	}

	championID, ok := c.ChampionMap[closestMatch]
	if !ok {
		return "", fmt.Errorf("champion ID not found for name '%s'", closestMatch)
	}

	return championID, nil
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
