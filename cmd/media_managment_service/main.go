package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/peano88/medias/config"
	"github.com/peano88/medias/internal/adapters/http"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("app", "media-management-service"),
		slog.String("version", "0.0.1"),
		slog.String("env", cfg.Env()),
	)
	logger.Info("Configuration loaded", slog.Any("configuration", cfg.AllSettings()))

	deps := http.Dependencies{
		TagCreator:      http.DummyTagCreator{},
		Logger:          logger,
		MetricForwarder: http.DummyMetricsForwarder{},
	}

	// Create server
	server := newServer(ctx, cfg, deps)

	// Run server with graceful shutdown
	if err := runServer(ctx, server, cfg, logger); err != nil {
		logger.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
