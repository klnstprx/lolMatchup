package data

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/models"
)

// DataLoader handles obtaining the latest patch and champion data, caching as needed.
type DataLoader struct {
	Config *config.AppConfig
	Client *client.Client
	Logger *log.Logger
	Cache  *cache.Cache
}

// NewDataLoader creates a DataLoader with references to config, client, and cache.
func NewDataLoader(cfg *config.AppConfig, client *client.Client, cache *cache.Cache) *DataLoader {
	return &DataLoader{
		Config: cfg,
		Client: client,
		Logger: cfg.Logger,
		Cache:  cache,
	}
}

// Initialize checks the latest patch from DDragon and refreshes champion data if needed.
func (dl *DataLoader) Initialize(ctx context.Context) error {
	cachedPatch := dl.Cache.GetPatch()

	latestPatch, err := dl.Client.FetchLatestPatch(ctx)
	if err != nil {
		// Fallback to cached patch if available (offline mode)
		if cachedPatch != "" {
			dl.Logger.Warnf("Could not fetch latest patch, using cached patch %s: %v", cachedPatch, err)
			dl.Config.PatchNumber = cachedPatch
			return nil
		}
		return fmt.Errorf("failed to fetch latest patch: %w", err)
	}
	dl.Logger.Infof("Latest patch version: %s", latestPatch)
	dl.Config.PatchNumber = latestPatch

	if cachedPatch != latestPatch {
		dl.Logger.Infof("Patch changed from %s to %s; invalidating cache.", cachedPatch, latestPatch)
		dl.Cache.Invalidate()
		dl.Cache.SetPatch(latestPatch)

		champions, err := dl.Client.FetchChampionList(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch champion map: %w", err)
		}
		nameMap, keyMap := buildChampionMaps(champions)
		dl.Cache.SetChampionMap(nameMap)
		dl.Cache.SetChampionKeyMap(keyMap)

		spells, err := dl.Client.FetchSummonerSpells(ctx, latestPatch)
		if err != nil {
			dl.Logger.Errorf("Could not fetch summoner spells: %v", err)
		} else {
			dl.Cache.SetSummonerSpells(spells)
		}

		if err := dl.Cache.Save(); err != nil {
			dl.Logger.Errorf("Could not save cache: %v", err)
		}
	} else {
		dl.Logger.Info("Patch is up to date. Checking champion map in cache.")
		if dl.Cache.GetChampionMapLen() == 0 {
			dl.Logger.Info("Champion map is empty; fetching from Meraki.")
			champions, err := dl.Client.FetchChampionList(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch champion map: %w", err)
			}
			nameMap, keyMap := buildChampionMaps(champions)
			dl.Cache.SetChampionMap(nameMap)
			dl.Cache.SetChampionKeyMap(keyMap)

			if err := dl.Cache.Save(); err != nil {
				dl.Logger.Errorf("Could not save cache: %v", err)
			}
		}
		if dl.Cache.GetSummonerSpellsLen() == 0 {
			dl.Logger.Info("Summoner spells cache is empty; fetching from DDragon.")
			spells, err := dl.Client.FetchSummonerSpells(ctx, latestPatch)
			if err != nil {
				dl.Logger.Errorf("Could not fetch summoner spells: %v", err)
			} else {
				dl.Cache.SetSummonerSpells(spells)
				if err := dl.Cache.Save(); err != nil {
					dl.Logger.Errorf("Could not save cache: %v", err)
				}
			}
		}
	}

	return nil
}

// buildChampionMaps returns a name->key map and a numeric ID->key map from champion data.
func buildChampionMaps(champions map[string]models.Champion) (nameMap, keyMap map[string]string) {
	nameMap = make(map[string]string, len(champions))
	keyMap = make(map[string]string, len(champions))
	for _, champ := range champions {
		nameMap[champ.Name] = champ.Key
		keyMap[strconv.Itoa(champ.ID)] = champ.Key
	}
	return nameMap, keyMap
}
