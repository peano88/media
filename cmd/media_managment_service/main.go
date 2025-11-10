package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/peano88/medias/internal/adapters/http"
	"github.com/peano88/medias/internal/adapters/metrics/expvar"
	"github.com/peano88/medias/internal/adapters/storage/postgres"
	"github.com/peano88/medias/internal/app/createtag"
)

func main() {
	// Load configuration
	cfg, err := LoadConfig()
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
	logger.Info("Configuration loaded", slog.Any("configuration", cfg.ConfigLoader().AllSettings()))

	pool, err := postgres.NewPool(ctx, &cfg.Database)
	if err != nil {
		logger.Error("Failed to create database pool",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("Database connection pool established")

	// Create repositories
	tagRepo := postgres.NewTagRepository(pool)

	// Create use cases
	createTagUseCase := createtag.New(tagRepo)

	deps := http.Dependencies{
		TagCreator:      createTagUseCase,
		Logger:          logger,
		MetricForwarder: expvar.NewExpvarMetrics(),
	}

	// Create server
	server := newServer(ctx, &cfg.Server, deps)

	// Run server with graceful shutdown
	if err := runServer(ctx, server, &cfg.Server, logger); err != nil {
		logger.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
