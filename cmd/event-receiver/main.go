package main

import (
	"context"
	"event-receiver/internal/config"
	"event-receiver/internal/handlers"
	"event-receiver/internal/kafka"
	appmetrics "event-receiver/internal/metrics"
	"log/slog"
	"net/http"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"os"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func environmentInitialization() *config.Config {
	err := godotenv.Load()
	setupLogger()
	if err != nil {
		slog.Debug("No .env file found, using environment variables")
	}

	return config.Load()
}

func createApiHandler(kafkaProducer *kafka.Producer) http.Handler {
	apiHandler := handlers.NewAPIHandler(kafkaProducer)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", apiHandler.HealthCheck)
	mux.HandleFunc("POST /sources/{source_name}/events", apiHandler.PostEvent)
	mux.Handle("GET /metrics", promhttp.Handler())

	return httpMetricsMiddleware(loggingMiddleware(mux))
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func normalizePath(path string) string {
	if strings.HasPrefix(path, "/sources/") && strings.HasSuffix(path, "/events") {
		return "/sources/{source_name}/events"
	}
	return path
}

func httpMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)
		appmetrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rw.statusCode)).Inc()
		appmetrics.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

func main() {

	cfg := environmentInitialization()
	exitCode := make(chan int, 1)
	defer func() {
		select {
		case code := <-exitCode:
			os.Exit(code)
		default:
		}
	}()

	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		slog.Error(
			"Failed to create Kafka producer",
			slog.Any("error", err),
		)
		exitCode <- 1
		return
	}
	defer kafkaShutdown(kafkaProducer)

	slog.Info("Kafka producer initialized")

	apiHandler := createApiHandler(kafkaProducer)

	server := http.Server{
		Addr:         cfg.Ip + ":" + cfg.Port,
		Handler:      apiHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	shutdownSignal := make(chan error, 1)
	go func() {
		slog.Info("Server starting", slog.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			shutdownSignal <- err
		}
	}()
	defer httpServerShutdown(&server)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("Shutting down server...")
		exitCode <- 0
		return
	case err := <-shutdownSignal:
		if err != nil {
			slog.Error("Server error, initializing shutdown", slog.Any("error", err))
			exitCode <- 1
		} else {
			exitCode <- 0
		}
		return
	}
}

func httpServerShutdown(server *http.Server) {
	slog.Info("Shutting down http server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Fail to shutdown http server", slog.Any("error", err))
	} else {
		slog.Info("Http server exited gracefully")
	}
}

func kafkaShutdown(kafkaProducer *kafka.Producer) {
	slog.Info("Shutting down kafka producer...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := kafkaProducer.Close(ctx); err != nil {
		slog.Error("Fail to shutdown kafka producer", slog.Any("error", err))
	} else {
		slog.Info("Kafka prodcser exited gracefully")
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
