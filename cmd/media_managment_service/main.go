package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/peano88/medias/internal/adapters/filestorage/s3"
	"github.com/peano88/medias/internal/adapters/http"
	"github.com/peano88/medias/internal/adapters/metrics/expvar"
	"github.com/peano88/medias/internal/adapters/storage/postgres"
	"github.com/peano88/medias/internal/app/createmedia"
	"github.com/peano88/medias/internal/app/createtag"
	"github.com/peano88/medias/internal/app/finalizemedia"
	"github.com/peano88/medias/internal/app/getmedia"
	"github.com/peano88/medias/internal/app/gettags"
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
	mediaRepo := postgres.NewMediaRepository(pool)

	// Create file storage adapter
	mediaSaver, err := s3.NewMediaSaver(ctx, cfg.S3, logger)
	if err != nil {
		logger.Error("Failed to create S3 media saver",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	logger.Info("S3 media saver initialized")

	// Create use cases
	createTagUseCase := createtag.New(tagRepo)
	getTagsUseCase := gettags.New(tagRepo)
	createMediaUseCase := createmedia.New(mediaRepo, mediaSaver)
	finalizeMediaUseCase := finalizemedia.New(mediaRepo, mediaSaver)
	getMediaUseCase := getmedia.New(mediaRepo, mediaSaver)

	deps := http.Dependencies{
		TagCreator:      createTagUseCase,
		TagRetriever:    getTagsUseCase,
		MediaCreator:    createMediaUseCase,
		MediaFinalizer:  finalizeMediaUseCase,
		MediaRetriever:  getMediaUseCase,
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
