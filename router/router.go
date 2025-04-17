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
	homeHandler := handlers.NewHomeHandler(cfg, apiClient)
	autocompleteHandler := handlers.NewAutocompleteHandler(cfg, apiClient)

	r.GET("/", homeHandler.HomeGET)
	r.GET("/champion", championHandler.ChampionGET)
	r.GET("/autocomplete", autocompleteHandler.AutocompleteGET)

	return r
}
