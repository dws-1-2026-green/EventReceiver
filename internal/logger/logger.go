package logger

import (
	"event-receiver/internal/config"
	"log/slog"
	"os"
	"strings"
)

func getLevel(config *config.Config) slog.Level {
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getHandler(config *config.Config, opts *slog.HandlerOptions) slog.Handler {
	if strings.ToLower(config.LogFormat) == "json" {
		return slog.NewJSONHandler(os.Stdout, opts)
	} else {
		return slog.NewTextHandler(os.Stdout, opts)
	}
}

func SetupLogger(cfg *config.Config) {
	var level slog.Level = getLevel(cfg)

	hostname, err := os.Hostname()

	if err != nil {
		hostname = "unknown"
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler = getHandler(cfg, opts)

	logger := slog.New(handler).With(
		slog.String("service", "event-receiver"),
		slog.String("hostname", hostname),
	)

	slog.SetDefault(logger)

	slog.Info("Logger initialized",
		slog.String("level", level.String()),
		slog.String("format", cfg.LogFormat),
	)

	if err != nil {
		slog.Warn("Failed to get hostname, using 'unknown'", slog.String("error", err.Error()))
	}

}
