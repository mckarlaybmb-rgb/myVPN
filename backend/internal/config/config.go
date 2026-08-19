package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
	APIKey      string
	XrayAPIAddr string
	XrayInbound string
}

func Load() Config {
	return Config{
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://vpnos:vpnos_pass@localhost:5432/vpnos?sslmode=disable"),
		Port:        envOrDefault("BACKEND_PORT", "8080"),
		APIKey:      envOrDefault("ADMIN_API_KEY", "CHANGE_ME"),
		XrayAPIAddr: envOrDefault("XRAY_API_ADDR", "127.0.0.1:12789"),
		XrayInbound: envOrDefault("XRAY_INBOUND_TAG", "vless-reality-tcp"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
