package router

import (
	"github.com/gin-gonic/gin"
	"github.com/klnstprx/lolMatchup/handlers"
	"github.com/klnstprx/lolMatchup/middleware"
	"github.com/klnstprx/lolMatchup/renderer"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// Load templates and static files if needed
	// r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	ginHtmlRenderer := r.HTMLRender
	r.HTMLRender = &renderer.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// Routes

	r.GET("/", handlers.HomeGET)
	r.GET("/champion", handlers.ChampionGET)

	return r
}
