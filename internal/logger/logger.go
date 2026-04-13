package logger

import (
	"log/slog"
	"os"
)

// Setup configures the global slog logger.
// In production (LOG_FORMAT=json), it outputs structured JSON.
// Otherwise, it uses human-readable text format.
func Setup() {
	var handler slog.Handler

	format := os.Getenv("LOG_FORMAT")
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	slog.SetDefault(slog.New(handler))
}
