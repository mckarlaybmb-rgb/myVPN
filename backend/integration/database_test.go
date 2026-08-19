package integration

import (
	"context"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/database"
	"os"
	"testing"
)

func TestPostgresMigrations(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	pool, err := database.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := database.Migrate(context.Background(), pool, "../migrations"); err != nil {
		t.Fatal(err)
	}
}
