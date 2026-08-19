package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
	APIKey      string
}

func Load() Config {
	return Config{
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://vpnos:vpnos_pass@localhost:5432/vpnos?sslmode=disable"),
		Port:        envOrDefault("BACKEND_PORT", "8080"),
		APIKey:      envOrDefault("ADMIN_API_KEY", "CHANGE_ME"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
