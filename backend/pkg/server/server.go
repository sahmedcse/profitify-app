package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"profitify-backend/pkg/config"
	"syscall"

	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	log        *zap.SugaredLogger
}

func New(handler http.Handler, cfg *config.Config, log *zap.SugaredLogger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + cfg.Port,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		config: cfg,
		log:    log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	serverErrors := make(chan error, 1)

	go func() {
		s.log.Infow("starting server",
			"port", s.config.Port,
			"environment", s.config.Environment,
		)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-shutdown:
		s.log.Infow("shutdown signal received", "signal", sig.String())
		return s.gracefulShutdown(ctx)
	case <-ctx.Done():
		s.log.Infow("context cancelled, shutting down")
		return s.gracefulShutdown(ctx)
	}

	return nil
}

func (s *Server) gracefulShutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	s.log.Infow("attempting graceful shutdown", "timeout", s.config.ShutdownTimeout)
	
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.httpServer.Close()
		return err
	}

	s.log.Info("server stopped successfully")
	return nil
}