package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	Environment     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
