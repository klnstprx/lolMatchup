package data

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
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
	latestPatch, err := dl.Client.FetchLatestPatch(ctx)
	if err != nil {
		// Fallback to cached patch if available (offline mode)
		if dl.Cache.Patch != "" {
			dl.Logger.Warnf("Could not fetch latest patch, using cached patch %s: %v", dl.Cache.Patch, err)
			dl.Config.PatchNumber = dl.Cache.Patch
			return nil
		}
		return fmt.Errorf("failed to fetch latest patch: %w", err)
	}
	dl.Logger.Infof("Latest patch version: %s", latestPatch)
	dl.Config.PatchNumber = latestPatch

	if dl.Cache.Patch != latestPatch {
		dl.Logger.Infof("Patch changed from %s to %s; invalidating cache.", dl.Cache.Patch, latestPatch)
		dl.Cache.Invalidate()
		dl.Cache.Patch = latestPatch

		champions, err := dl.Client.FetchChampionList(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch champion map: %w", err)
		}
		nameMap := make(map[string]string)
		for _, champ := range champions {
			nameMap[champ.Name] = champ.Key
		}
		dl.Cache.SetChampionMap(nameMap)

		if err := dl.Cache.Save(); err != nil {
			dl.Logger.Errorf("Could not save cache: %v", err)
		}
	} else {
		dl.Logger.Info("Patch is up to date. Checking champion map in cache.")
		if len(dl.Cache.ChampionMap) == 0 {
			dl.Logger.Info("Champion map is empty; fetching from Meraki.")
			champions, err := dl.Client.FetchChampionList(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch champion map: %w", err)
			}
			nameMap := make(map[string]string)
			for _, champ := range champions {
				nameMap[champ.Name] = champ.Key
			}
			dl.Cache.SetChampionMap(nameMap)

			if err := dl.Cache.Save(); err != nil {
				dl.Logger.Errorf("Could not save cache: %v", err)
			}
		}
	}

	return nil
}
