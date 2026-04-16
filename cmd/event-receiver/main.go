package main

import (
	"event-receiver/internal/config"
	"event-receiver/internal/handlers"
	"event-receiver/internal/kafka"
	"log/slog"
	"net/http"

	"os"
	"strings"

	"github.com/joho/godotenv"
)

func setupLogger() {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}

	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))

	slog.Info("Logger initialized",
		slog.String("level", levelStr),
		slog.String("format", os.Getenv("LOG_FORMAT")),
	)
}

func main() {
	err := godotenv.Load()
	setupLogger()

	if err != nil {
		slog.Debug("No .env file found, using environment variables")
	}

	cfg := config.Load()

	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		slog.Error(
			"Failed to create Kafka producer",
			slog.Any("error", err),
		)
		panic(err)
	}

	defer kafkaProducer.Close()
	slog.Info("Kafka producer initialized")

	apiHandler := handlers.NewAPIHandler(kafkaProducer)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", apiHandler.HealthCheck)
	mux.HandleFunc("POST /sources/{source_name}/events", apiHandler.PostEvent)

	handlerWithMiddleware := loggingMiddleware(mux)

	addr := cfg.Ip + ":" + cfg.Port
	slog.Info(
		"Server starting",
		slog.String("addres", addr),
	)
	if err := http.ListenAndServe(addr, handlerWithMiddleware); err != nil {
		slog.Error(
			"Server failed to start",
			slog.Any("error", err),
		)
		panic(err)
	}

}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug(
			"Http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
		)
		next.ServeHTTP(w, r)
	})
}
