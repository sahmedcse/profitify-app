package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"profitify-backend/internal/middleware"
	"profitify-backend/pkg/logger"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
    logger.Init()
    log := logger.Get()
    ctx := context.Background()

    ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Info("Shutting down server...")
        cancel()
    }()


	router := gin.New()
    router.Use(gin.Recovery())
    router.Use(middleware.Log())

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