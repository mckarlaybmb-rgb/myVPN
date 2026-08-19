package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() Config {
	return Config{DatabaseURL: envOrDefault("DATABASE_URL", "postgres://vpnos:vpnos_pass@localhost:5432/vpnos?sslmode=disable"), Port: envOrDefault("BACKEND_PORT", "8080")}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" { return value }
	return fallback
}