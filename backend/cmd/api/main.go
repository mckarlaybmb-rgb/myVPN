package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/config"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/database"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/handlers"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
)
func main() {
	cfg := config.Load(); ctx := context.Background(); pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil { log.Printf("database unavailable: %v", err); os.Exit(1) }; defer pool.Close()
	if err := database.Migrate(ctx, pool, "migrations"); err != nil { log.Fatal(err) }
	queue := jobs.NewQueue(pool); routes := handlers.New(services.NewUserService(pool), services.NewSubscriptionService(pool, queue)); mux := http.NewServeMux()
	mux.HandleFunc("GET /health", routes.Health); mux.HandleFunc("/api/v1/users", routes.Users); mux.HandleFunc("/api/v1/subscriptions", routes.Subscriptions)
	log.Printf("API listening on :%s", cfg.Port); if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil { log.Fatal(err) }
}