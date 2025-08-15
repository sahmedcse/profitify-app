package router

import (
	"profitify-backend/internal/handlers"
	"profitify-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func New(mode string) *Router {
	if mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Log())

	return &Router{
		engine: r,
	}
}

func (r *Router) SetupRoutes(handler *handlers.Handler) {
	r.setupHealthRoutes()
	r.setupAPIRoutes(handler)
}

func (r *Router) setupHealthRoutes() {
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/health/live", r.livenessCheck)
	r.engine.GET("/health/ready", r.readinessCheck)
}

func (r *Router) setupAPIRoutes(handler *handlers.Handler) {
	api := r.engine.Group("/api")
	{
		api.GET("/tickers", handler.GetAllTickers)
	}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "profitify-backend",
	})
}

func (r *Router) livenessCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "alive",
	})
}

func (r *Router) readinessCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ready",
	})
}