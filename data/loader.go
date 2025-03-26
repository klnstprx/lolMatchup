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
	latestPatch, err := dl.Client.FetchLatestPatch(ctx, dl.Config.DDragonVersionURL)
	if err != nil {
		return fmt.Errorf("failed to fetch latest patch: %w", err)
	}
	dl.Logger.Infof("Latest patch version: %s", latestPatch)
	dl.Config.PatchNumber = latestPatch

	if dl.Cache.Patch != latestPatch {
		dl.Logger.Infof("Patch changed from %s to %s; invalidating cache.", dl.Cache.Patch, latestPatch)
		dl.Cache.Invalidate()
		dl.Cache.Patch = latestPatch

		championMap, err := dl.Client.FetchChampionNameIDMap(ctx,
			dl.Config.DDragonURL, latestPatch, dl.Config.LanguageCode)
		if err != nil {
			return fmt.Errorf("failed to fetch champion map: %w", err)
		}
		dl.Cache.SetChampionMap(championMap)

		if err := dl.Cache.Save(); err != nil {
			dl.Logger.Errorf("Could not save cache: %v", err)
		}
	} else {
		dl.Logger.Info("Patch is up to date. Checking champion map in cache.")
		if len(dl.Cache.ChampionMap) == 0 {
			dl.Logger.Info("Champion map is empty; fetching from Data Dragon.")
			championMap, err := dl.Client.FetchChampionNameIDMap(ctx,
				dl.Config.DDragonURL, latestPatch, dl.Config.LanguageCode)
			if err != nil {
				return fmt.Errorf("failed to fetch champion map: %w", err)
			}
			dl.Cache.SetChampionMap(championMap)

			if err := dl.Cache.Save(); err != nil {
				dl.Logger.Errorf("Could not save cache: %v", err)
			}
		}
	}

	dl.Config.SetDDragonDataURL()
	return nil
}
