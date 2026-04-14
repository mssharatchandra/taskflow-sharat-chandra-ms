package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/config"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/database"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/logger"
	"github.com/sharatchandra/taskflow-sharat-chandra-ms/internal/server"
)

func main() {
	// Set up structured logging
	logger.Setup()

	// Load .env file if present (ignored in production/Docker)
	_ = godotenv.Load()

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to the database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("connected to database")

	// Create and start the server
	srv := server.New(cfg, db)

	// Graceful shutdown: listen for termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Block until we receive a signal
	sig := <-quit
	slog.Info("received shutdown signal", "signal", sig.String())

	// Give in-flight requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
