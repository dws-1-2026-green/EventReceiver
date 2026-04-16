package main

import (
	"event-receiver/config"
	"event-receiver/handlers"
	"event-receiver/kafka"
	"log/slog"
	"net/http"
)

func main() {
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
