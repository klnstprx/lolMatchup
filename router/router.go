package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/handlers"
	"github.com/klnstprx/lolMatchup/middleware"
	"github.com/klnstprx/lolMatchup/renderer"
	"github.com/klnstprx/lolMatchup/static"
	"golang.org/x/time/rate"
)

// SetupRouter configures Gin, applying custom renderer and middleware,
// then registers routes.
func SetupRouter(cfg *config.AppConfig, apiClient *client.Client) *gin.Engine {
	r := gin.New()

	// Middlewares: request ID first (so it's available to the logger), then logging and recovery
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(cfg.Logger))
	r.Use(middleware.RecoveryMiddleware(cfg.Logger))

	// Serve embedded static files under /static
	r.StaticFS("/static", http.FS(static.FS))

	// Wrap default gin HTML renderer in our custom templ renderer
	defaultGinRenderer := r.HTMLRender
	r.HTMLRender = &renderer.HTMLTemplRenderer{
		FallbackHtmlRenderer: defaultGinRenderer,
	}

	// Handlers
	championHandler := handlers.NewChampionHandler(cfg, apiClient)
	autocompleteHandler := handlers.NewAutocompleteHandler(cfg, apiClient)
	playerHandler := handlers.NewPlayerHandler(cfg, apiClient)
	liveGameHandler := handlers.NewLiveGameHandler(cfg, apiClient)
	matchHandler := handlers.NewMatchHandler(cfg, apiClient)
	pageHandler := handlers.NewPageHandler(cfg, championHandler, playerHandler)

	// Cache policies
	pageCache := middleware.CacheControl("public, max-age=300")
	championCache := middleware.CacheControl("public, max-age=3600")
	autocompleteCache := middleware.CacheControl("public, max-age=30")

	// Page routes
	r.GET("/", pageCache, pageHandler.HomePageGET)
	r.GET("/search", pageHandler.SearchGET)
	r.GET("/champion-search", pageCache, championHandler.ChampionPageGET)

	// Legacy redirects — player and livegame searches now go through the homepage
	r.GET("/player-search", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})
	r.GET("/livegame-search", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	// Champion data changes per patch — cache for 1 hour
	r.GET("/champion", championCache, championHandler.ChampionGET)
	r.GET("/autocomplete", autocompleteCache, autocompleteHandler.AutocompleteGET)

	// Routes that call Riot API — rate limited, no cache (real-time data)
	riotLimiter := middleware.RateLimitMiddleware(rate.Limit(15), 20)
	r.GET("/player", riotLimiter, playerHandler.PlayerGET)
	r.GET("/player/livegame", riotLimiter, liveGameHandler.PlayerLiveGameGET)
	r.GET("/livegame", riotLimiter, liveGameHandler.LiveGameGET)
	r.GET("/match", riotLimiter, matchHandler.MatchGET)
	r.GET("/match/player", riotLimiter, matchHandler.MatchPlayerGET)

	return r
}
