package main

import (
	"net/http"

	"profitify-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
    logger.Init()
    log := logger.Get()

	router := gin.New()
    router.Use(gin.Recovery())
    router.Use(logger.Middleware())

    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "OK"})
    })

    server := &http.Server{
        Addr: ":8080",
        Handler: router,
    }

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}