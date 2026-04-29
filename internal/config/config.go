package config

import (
	"os"
)

type Config struct {
	Ip           string
	Port         string
	KafkaBrokers []string
	KafkaTopic   string
	LogLevel     string
	LogFormat    string
}

func Load() *Config {
	return &Config{
		Ip:           getEnv("IP", "0.0.0.0"),
		Port:         getEnv("PORT", "8080"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:   getEnv("KAFKA_TOPIC", "routing.requests"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		LogFormat:    getEnv("LOG_FORMAT", "text"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
