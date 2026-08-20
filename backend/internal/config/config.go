package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL         string
	Port                string
	APIKey              string
	XUIBaseURL          string
	XUIUsername         string
	XUIPassword         string
	XUIInboundID        int64
	TelegramEnabled     bool
	TelegramBotToken    string
	TelegramAdminChatID int64
}

func Load() Config {
	return Config{
		DatabaseURL:         envOrDefault("DATABASE_URL", "postgres://vpnos:vpnos_pass@localhost:5432/vpnos?sslmode=disable"),
		Port:                envOrDefault("BACKEND_PORT", "8080"),
		APIKey:              envOrDefault("ADMIN_API_KEY", "CHANGE_ME"),
		XUIBaseURL:          envOrDefault("XUI_BASE_URL", ""),
		XUIUsername:         envOrDefault("XUI_USERNAME", ""),
		XUIPassword:         envOrDefault("XUI_PASSWORD", ""),
		XUIInboundID:        envInt64OrDefault("XUI_INBOUND_ID", 0),
		TelegramEnabled:     os.Getenv("TELEGRAM_ENABLED") == "true",
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramAdminChatID: envInt64OrDefault("TELEGRAM_ADMIN_CHAT_ID", 0),
	}
}

func (config Config) Validate() error {
	if config.DatabaseURL == "" || config.APIKey == "" || config.APIKey == "CHANGE_ME" {
		return fmt.Errorf("DATABASE_URL and a non-default ADMIN_API_KEY are required")
	}
	if config.TelegramEnabled && config.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required when Telegram is enabled")
	}
	return nil
}

func envInt64OrDefault(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	var parsed int64
	if _, err := fmt.Sscan(value, &parsed); err != nil {
		return fallback
	}
	return parsed
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
