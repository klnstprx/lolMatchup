package data

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/klnstprx/lolMatchup/cache"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
)

type DataLoader struct {
	Config *config.AppConfig
	Client *client.Client
	Logger *log.Logger
	Cache  *cache.Cache
}

func NewDataLoader(cfg *config.AppConfig, client *client.Client, cache *cache.Cache) *DataLoader {
	return &DataLoader{
		Config: cfg,
		Client: client,
		Logger: cfg.Logger,
		Cache:  cache,
	}
}

// Initialize checks for the latest patch and updates champion data if necessary
func (dl *DataLoader) Initialize(ctx context.Context) error {
	// Fetch the latest patch number
	latestPatch, err := dl.Client.FetchLatestPatch(ctx, dl.Config.DDragonVersionURL)
	if err != nil {
		return err
	}
	dl.Logger.Infof("Latest patch version: %s", latestPatch)
	dl.Config.PatchNumber = latestPatch

	cachePatchVersion := dl.Cache.Patch
	if cachePatchVersion != latestPatch {
		dl.Logger.Infof("Patch version changed from %s to %s. Invalidating cache.", cachePatchVersion, latestPatch)
		// Invalidate the cache
		dl.Cache.Invalidate()
		dl.Cache.Patch = latestPatch

		// Fetch the new champion map
		championMap, err := dl.Client.FetchChampionNameIDMap(ctx, dl.Config.DDragonURL, latestPatch, dl.Config.LanguageCode)
		if err != nil {
			return err
		}
		// Update the cache with the new champion map
		dl.Cache.SetChampionMap(championMap)

		// Save the updated cache to disk
		if err := dl.Cache.Save(); err != nil {
			dl.Logger.Errorf("Error saving cache: %v", err)
		}
	} else {
		dl.Logger.Info("Patch is up to date. Loading champion map from cache.")
		// If the ChampionMap is empty, we might need to fetch it
		if len(dl.Cache.ChampionMap) == 0 {
			dl.Logger.Info("Champion map is empty. Fetching champion map.")
			championMap, err := dl.Client.FetchChampionNameIDMap(ctx, dl.Config.DDragonURL, latestPatch, dl.Config.LanguageCode)
			if err != nil {
				return err
			}
			// Update the cache with the fetched champion map
			dl.Cache.SetChampionMap(championMap)

			// Save the updated cache to disk
			if err := dl.Cache.Save(); err != nil {
				dl.Logger.Errorf("Error saving cache: %v", err)
			}
		}
	}

	dl.Config.SetDDragonDataURL()

	return nil
}
