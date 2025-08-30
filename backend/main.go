package main

import (
	"context"
	"fmt"
	"os"
	"profitify-backend/internal/handlers"
	"profitify-backend/pkg/config"
	"profitify-backend/pkg/logger"
	"profitify-backend/pkg/router"
	"profitify-backend/pkg/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "startup failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Create root context for the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration first
	cfg := config.Load()

	// Initialize logger with configuration
	if err := logger.Init(&logger.Config{
		Level:       os.Getenv("LOG_LEVEL"),
		Environment: cfg.Environment,
		OutputPaths: []string{"stdout"},
	}); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	log := logger.Get()
	defer func() {
		_ = logger.Sync()
	}()

	// Initialize router
	r := router.New(cfg.Environment)

	// Initialize handlers with application context
	handler, err := handlers.NewHandler(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize handlers: %w", err)
	}

	// Setup routes
	r.SetupRoutes(handler)

	// Create and start server with context
	srv := server.New(r.Engine(), cfg, log)
	return srv.Start(ctx)
}
