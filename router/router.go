package router

import (
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/client"
	"github.com/klnstprx/lolMatchup/config"
	"github.com/klnstprx/lolMatchup/handlers"
	"github.com/klnstprx/lolMatchup/middleware"
	"github.com/klnstprx/lolMatchup/renderer"
)

func SetupRouter(cfg *config.AppConfig, apiClient *client.Client) *gin.Engine {
	r := gin.New()

	// Set up middleware with cfg.Logger
	r.Use(middleware.LoggerMiddleware(cfg.Logger))
	r.Use(middleware.RecoveryMiddleware(cfg.Logger))

	// Serve static files
	r.Static("/static", "./static")
	ginHtmlRenderer := r.HTMLRender
	r.HTMLRender = &renderer.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// Initialize handlers with dependencies
	championHandler := handlers.NewChampionHandler(cfg, apiClient)
	homeHandler := handlers.NewHomeHandler(cfg, apiClient)

	// Routes
	r.GET("/", homeHandler.HomeGET)
	r.GET("/champion", championHandler.ChampionGET)

	return r
}
