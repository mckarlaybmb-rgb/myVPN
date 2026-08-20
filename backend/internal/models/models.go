package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type XrayClient struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	Email           string         `json:"email"`
	UUID            string         `json:"uuid"`
	SubscriptionURL string         `json:"subscription_url"`
	Protocol        string         `json:"protocol"`
	Config          map[string]any `json:"config"`
	Enabled         bool           `json:"enabled"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type Subscription struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
