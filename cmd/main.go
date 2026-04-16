package main

import (
	"event-receiver/config"
	"event-receiver/handlers"
	"event-receiver/kafka"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	defer kafkaProducer.Close()
	log.Println("Kafka producer initialized")

	apiHandler := handlers.NewAPIHandler(kafkaProducer)

	router := mux.NewRouter()

	router.Use(loggingMiddleware)
	router.Use(contentTypeMiddleware)
	router.HandleFunc("/health", apiHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/api/event", apiHandler.PostEvent).Methods("POST")

	addr := cfg.Ip + ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
