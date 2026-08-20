package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/config"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/database"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/handlers"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/middleware"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/repositories"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/telegram"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/xray"
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
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
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
	xrayService := xray.NewService(repositories.NewXrayClientRepository(pool), xray.NewClient(xray.Config{BaseURL: cfg.XUIBaseURL, Username: cfg.XUIUsername, Password: cfg.XUIPassword, InboundID: cfg.XUIInboundID}), "x-ui")
	routes := handlers.New(services.NewUserService(repositories.NewUserRepository(pool), xrayService), services.NewSubscriptionService(repositories.NewSubscriptionRepository(pool), queue, xrayService), repositories.NewAdminRepository(pool))
	worker := jobs.NewWorker(queue, time.Second, map[string]jobs.Handler{"subscription.created": func(context.Context, jobs.Job) error { return nil }, "subscription.renewed": func(context.Context, jobs.Job) error { return nil }})
	jobContext, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		if err := worker.Run(jobContext); err != nil && jobContext.Err() == nil {
			log.Printf("job worker stopped: %v", err)
		}
	}()
	adminRepository := repositories.NewAdminRepository(pool)
	go func() { _ = jobs.NewNodeChecker(adminRepository).Run(jobContext) }()
	var notifier jobs.Notifier
	if cfg.TelegramEnabled {
		store := telegram.NewPGStore(pool)
		bot := telegram.NewBot(cfg.TelegramBotToken, store)
		notifier = telegram.NewNotifier(pool, bot)
		go func() {
			if err := bot.Poll(jobContext); err != nil && jobContext.Err() == nil {
				log.Printf("telegram bot stopped: %v", err)
			}
		}()
	}
	go func() {
		_ = jobs.NewExpiryScheduler(repositories.NewSubscriptionRepository(pool), xrayService, notifier).Run(jobContext)
	}()

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		// User routes
		switch {
		case method == http.MethodPost && path == "/api/v1/users":
			routes.CreateUser(w, r)
			return
		case method == http.MethodGet && path == "/api/v1/admin/stats":
			routes.AdminStats(w, r)
			return
		case method == http.MethodGet && path == "/api/v1/admin/nodes":
			routes.AdminNodes(w, r)
			return
		case method == http.MethodGet && path == "/api/v1/admin/subscriptions":
			routes.AdminSubscriptions(w, r)
			return
		case method == http.MethodGet && path == "/api/v1/admin/users":
			routes.AdminUsers(w, r)
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
		case method == http.MethodPost && strings.HasPrefix(path, "/api/v1/subscriptions/") && strings.HasSuffix(path, "/suspend"):
			id := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/subscriptions/"), "/suspend")
			r.SetPathValue("id", strings.TrimSuffix(id, "/"))
			routes.SuspendSubscription(w, r)
			return
		case method == http.MethodPost && strings.HasPrefix(path, "/api/v1/subscriptions/") && strings.HasSuffix(path, "/expire"):
			id := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/subscriptions/"), "/expire")
			r.SetPathValue("id", strings.TrimSuffix(id, "/"))
			routes.ExpireSubscription(w, r)
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
