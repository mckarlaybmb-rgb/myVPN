package integration

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/database"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"os"
	"testing"
	"time"
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
