package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/config"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/database"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/handlers"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/middleware"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/repositories"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
)

func resolveMigrationsDir() string {
	// Try a few sensible locations relative to this file and the working dir
	_, file, _, ok := runtime.Caller(0)
	if ok {
		base := filepath.Dir(file) // backend/cmd/api
		candidates := []string{
			filepath.Join(base, "../../migrations"), // backend/migrations
			filepath.Join(base, "../migrations"),
			"./migrations",
			"migrations",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}
	return "migrations"
}

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("database unavailable: %v", err)
		os.Exit(1)
	}
	defer pool.Close()
	migrationsDir := resolveMigrationsDir()
	if err := database.Migrate(ctx, pool, migrationsDir); err != nil {
		log.Fatal(err)
	}
	queue := jobs.NewQueue(pool)
	routes := handlers.New(services.NewUserService(repositories.NewUserRepository(pool)), services.NewSubscriptionService(repositories.NewSubscriptionRepository(pool), queue))

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		// User routes
		switch {
		case method == http.MethodPost && path == "/api/v1/users":
			routes.CreateUser(w, r)
			return
		case method == http.MethodGet && path == "/api/v1/users":
			routes.ListUsers(w, r)
			return
		case method == http.MethodDelete && strings.HasPrefix(path, "/api/v1/users/"):
			id := strings.TrimPrefix(path, "/api/v1/users/")
			r.SetPathValue("id", id)
			routes.DeleteUser(w, r)
			return

		// Subscription routes
		case method == http.MethodPost && path == "/api/v1/subscriptions":
			routes.CreateSubscription(w, r)
			return
		case method == http.MethodGet && strings.HasPrefix(path, "/api/v1/subscriptions/") && !strings.HasSuffix(path, "/renew"):
			userID := strings.TrimPrefix(path, "/api/v1/subscriptions/")
			r.SetPathValue("user_id", userID)
			routes.ListSubscriptions(w, r)
			return
		case method == http.MethodPost && strings.HasPrefix(path, "/api/v1/subscriptions/") && strings.HasSuffix(path, "/renew"):
			// path like /api/v1/subscriptions/{id}/renew
			trim := strings.TrimPrefix(path, "/api/v1/subscriptions/")
			id := strings.TrimSuffix(trim, "/renew")
			id = strings.TrimSuffix(id, "/")
			r.SetPathValue("id", id)
			routes.RenewSubscription(w, r)
			return
		default:
			http.NotFound(w, r)
			return
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", routes.Health)
	mux.Handle("/api/", middleware.APIKey(cfg.APIKey)(apiHandler))

	log.Printf("API listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
