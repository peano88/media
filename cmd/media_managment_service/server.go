package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	nethttp "net/http"
	"time"

	"github.com/peano88/medias/config"
	"github.com/peano88/medias/internal/adapters/http"
)

// newServer creates and configures a new HTTP server
func newServer(ctx context.Context, cfg *config.Config, deps http.Dependencies) *nethttp.Server {
	r := http.NewRouter(deps)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	server := &nethttp.Server{
		Addr:              addr,
		Handler:           r,
		ReadTimeout:       0, // TODO
		ReadHeaderTimeout: 0, // TODO
		WriteTimeout:      0, // TODO
		IdleTimeout:       0, // TODO
		MaxHeaderBytes:    0, // TODO
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	return server
}

// runServer starts the HTTP server and handles graceful shutdown
func runServer(ctx context.Context, server *nethttp.Server, cfg *config.Config, logger *slog.Logger) error {
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Info("Starting server", slog.String("address", server.Addr))

		if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			logger.Error("Server failed", slog.String("error", err.Error()))
			errChan <- err
		}
	}()

	// Wait for context cancellation (signal)
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	logger.Info("Shutting down server")

	// Create shutdown context with timeout
	shutDownCtx, cancelShutDownCtx := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second)
	defer cancelShutDownCtx()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutDownCtx); err != nil {
		logger.Error("Server shutdown failed", slog.String("error", err.Error()))
		return err
	}

	logger.Info("Server gracefully stopped")
	return nil
}
