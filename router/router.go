package router

import (
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/handlers"
	"github.com/klnstprx/lolMatchup/middleware"
	"github.com/klnstprx/lolMatchup/renderer"
)

// SetupRouter configures Gin, applying custom renderer and middleware,
// then registers routes.
func SetupRouter(cfg *config.AppConfig, apiClient *client.Client) *gin.Engine {
	r := gin.New()

	// Middlewares for logging and recovery
	r.Use(middleware.LoggerMiddleware(cfg.Logger))
	r.Use(middleware.RecoveryMiddleware(cfg.Logger))

	// Serve static files
	r.Static("/static", "./static")

	// Wrap default gin HTML renderer in our custom templ renderer
	defaultGinRenderer := r.HTMLRender
	r.HTMLRender = &renderer.HTMLTemplRenderer{
		FallbackHtmlRenderer: defaultGinRenderer,
	}

	// Handlers
	championHandler := handlers.NewChampionHandler(cfg, apiClient)
	homeHandler := handlers.NewHomeHandler(cfg, apiClient)
	autocompleteHandler := handlers.NewAutocompleteHandler(cfg, apiClient)

	r.GET("/", homeHandler.HomeGET)
	r.GET("/champion", championHandler.ChampionGET)
	r.GET("/autocomplete", autocompleteHandler.AutocompleteGET)

	return r
}
