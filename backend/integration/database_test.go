package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/database"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/repositories"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/xray"
)

type integrationRuntime struct{ disabled int }

func (runtime *integrationRuntime) AddUser(context.Context, models.XrayClient) (string, error) {
	return "https://x-ui.example/sub/client", nil
}
func (runtime *integrationRuntime) RemoveUser(context.Context, models.XrayClient) error   { return nil }
func (runtime *integrationRuntime) EnableClient(context.Context, models.XrayClient) error { return nil }
func (runtime *integrationRuntime) DisableClient(context.Context, models.XrayClient) error {
	runtime.disabled++
	return nil
}

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

func TestQueueTransitions(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool, "../migrations"); err != nil {
		t.Fatal(err)
	}

	var entityID string
	if err := pool.QueryRow(ctx, `INSERT INTO users (email) VALUES ($1) RETURNING id::text`, fmt.Sprintf("queue-%d@example.com", time.Now().UnixNano())).Scan(&entityID); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM job_queue WHERE entity_id = $1`, entityID)
	defer pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, entityID)

	queue := jobs.NewQueue(pool)
	completedID, err := queue.Enqueue(ctx, "test.completed", entityID, map[string]string{"case": "completed"})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := queue.Claim(ctx)
	if err != nil || claimed.ID != completedID {
		t.Fatalf("claim completed case: job=%#v err=%v", claimed, err)
	}
	if err := assertJobStatus(ctx, pool, completedID, jobs.Processing); err != nil {
		t.Fatal(err)
	}
	if err := queue.Complete(ctx, completedID); err != nil {
		t.Fatal(err)
	}
	if err := assertJobStatus(ctx, pool, completedID, jobs.Completed); err != nil {
		t.Fatal(err)
	}

	failedID, err := queue.Enqueue(ctx, "test.failed", entityID, map[string]string{"case": "failed"})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err = queue.Claim(ctx)
	if err != nil || claimed.ID != failedID {
		t.Fatalf("claim failed case: job=%#v err=%v", claimed, err)
	}
	if err := queue.Fail(ctx, failedID, "test failure"); err != nil {
		t.Fatal(err)
	}
	if err := assertJobStatus(ctx, pool, failedID, jobs.Failed); err != nil {
		t.Fatal(err)
	}
}

func TestXrayClientPersistenceAndSubscriptionDisable(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool, "../migrations"); err != nil {
		t.Fatal(err)
	}

	email := fmt.Sprintf("xray-%d@example.com", time.Now().UnixNano())
	user, err := repositories.NewUserRepository(pool).Create(ctx, email)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
	runtime := &integrationRuntime{}
	clientService := xray.NewService(repositories.NewXrayClientRepository(pool), runtime, "vless-reality-tcp")
	client, err := clientService.CreateClient(ctx, user.ID, user.Email)
	if err != nil {
		t.Fatal(err)
	}
	if client.UUID == "" || client.Protocol != "vless" || client.Config["inbound_tag"] != "vless-reality-tcp" {
		t.Fatalf("unexpected client: %#v", client)
	}
	var storedUUID string
	var storedConfig []byte
	if err := pool.QueryRow(ctx, `SELECT uuid::text, config FROM xray_clients WHERE user_id = $1`, user.ID).Scan(&storedUUID, &storedConfig); err != nil {
		t.Fatal(err)
	}
	if storedUUID != client.UUID || len(storedConfig) == 0 {
		t.Fatalf("stored uuid/config: %q %s", storedUUID, storedConfig)
	}

	queue := jobs.NewQueue(pool)
	subscription, err := services.NewSubscriptionService(repositories.NewSubscriptionRepository(pool), queue, clientService).Create(ctx, user.ID, "monthly", time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := services.NewSubscriptionService(repositories.NewSubscriptionRepository(pool), queue, clientService).Suspend(ctx, subscription.ID); err != nil {
		t.Fatal(err)
	}
	var enabled bool
	if err := pool.QueryRow(ctx, `SELECT enabled FROM xray_clients WHERE user_id = $1`, user.ID).Scan(&enabled); err != nil {
		t.Fatal(err)
	}
	if enabled || runtime.disabled != 1 {
		t.Fatalf("enabled=%v runtime=%#v", enabled, runtime)
	}
}

func assertJobStatus(ctx context.Context, pool *pgxpool.Pool, id string, expected jobs.Status) error {
	var status jobs.Status
	if err := pool.QueryRow(ctx, `SELECT status FROM job_queue WHERE id = $1`, id).Scan(&status); err != nil {
		return err
	}
	if status != expected {
		return fmt.Errorf("job %s has status %q, expected %q", id, status, expected)
	}
	return nil
}
