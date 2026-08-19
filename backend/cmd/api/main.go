package main

import (
	"context"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/config"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/database"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/handlers"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/middleware"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/repositories"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("database unavailable: %v", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool, "migrations"); err != nil {
		log.Fatal(err)
	}
	queue := jobs.NewQueue(pool)
	routes := handlers.New(services.NewUserService(repositories.NewUserRepository(pool)), services.NewSubscriptionService(repositories.NewSubscriptionRepository(pool), queue))
	api := http.NewServeMux()
	api.HandleFunc("POST /api/v1/users", routes.CreateUser)
	api.HandleFunc("GET /api/v1/users", routes.ListUsers)
	api.HandleFunc("DELETE /api/v1/users/{id}", routes.DeleteUser)
	api.HandleFunc("POST /api/v1/subscriptions", routes.CreateSubscription)
	api.HandleFunc("GET /api/v1/subscriptions/{user_id}", routes.ListSubscriptions)
	api.HandleFunc("POST /api/v1/subscriptions/{id}/renew", routes.RenewSubscription)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", routes.Health)
	mux.Handle("/api/", middleware.APIKey(cfg.APIKey)(api))
	log.Printf("API listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
