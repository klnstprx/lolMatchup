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
)

// SetupRouter configures Gin, applying custom renderer and middleware,
// then registers routes.
func SetupRouter(cfg *config.AppConfig, apiClient *client.Client) *gin.Engine {
	r := gin.New()

	// Middlewares for logging and recovery
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
	pageHandler := handlers.NewPageHandler(cfg)

	// Page routes
	r.GET("/", pageHandler.HomePageGET)
	r.GET("/champion-search", pageHandler.ChampionPageGET)
	r.GET("/player-search", pageHandler.PlayerPageGET)
	r.GET("/livegame-search", pageHandler.LiveGamePageGET)

	// AJAX/functional routes
	r.GET("/champion", championHandler.ChampionGET)
	r.GET("/autocomplete", autocompleteHandler.AutocompleteGET)
	r.GET("/player", playerHandler.PlayerGET)
	r.GET("/livegame", liveGameHandler.LiveGameGET)

	return r
}
