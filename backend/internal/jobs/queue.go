package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Status string

const (
	Pending    Status = "pending"
	Processing Status = "processing"
	Completed  Status = "completed"
	Failed     Status = "failed"
)

type Job struct {
	ID       string
	Type     string
	EntityID string
	Payload  []byte
}

type Queue struct{ pool *pgxpool.Pool }

func NewQueue(pool *pgxpool.Pool) *Queue { return &Queue{pool: pool} }
func (queue *Queue) Enqueue(ctx context.Context, jobType, entityID string, payload map[string]string) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	var id string
	err = queue.pool.QueryRow(ctx, `INSERT INTO job_queue (job_type, entity_id, payload) VALUES ($1, $2, $3) RETURNING id::text`, jobType, entityID, data).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("enqueue job: %w", err)
	}
	return id, nil
}

func (queue *Queue) Claim(ctx context.Context) (Job, error) {
	transaction, err := queue.pool.Begin(ctx)
	if err != nil {
		return Job{}, err
	}
	defer transaction.Rollback(ctx)
	var job Job
	err = transaction.QueryRow(ctx, `WITH next_job AS (SELECT id FROM job_queue WHERE status = 'pending' AND available_at <= NOW() ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1) UPDATE job_queue SET status = 'processing', attempts = attempts + 1 WHERE id = (SELECT id FROM next_job) RETURNING id::text, job_type, COALESCE(entity_id::text, ''), payload`).Scan(&job.ID, &job.Type, &job.EntityID, &job.Payload)
	if err != nil {
		return Job{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (queue *Queue) Complete(ctx context.Context, id string) error {
	_, err := queue.pool.Exec(ctx, `UPDATE job_queue SET status = 'completed', processed_at = NOW() WHERE id = $1 AND status = 'processing'`, id)
	return err
}

func (queue *Queue) Fail(ctx context.Context, id, message string) error {
	_, err := queue.pool.Exec(ctx, `UPDATE job_queue SET status = CASE WHEN attempts < 3 THEN 'pending' ELSE 'failed' END, available_at = CASE WHEN attempts < 3 THEN NOW() + INTERVAL '1 minute' ELSE available_at END, last_error = $2, processed_at = CASE WHEN attempts < 3 THEN NULL ELSE NOW() END WHERE id = $1 AND status = 'processing'`, id, message)
	return err
}

type Handler func(context.Context, Job) error
type Worker struct {
	queue        *Queue
	handlers     map[string]Handler
	pollInterval time.Duration
}

func NewWorker(queue *Queue, pollInterval time.Duration, handlers map[string]Handler) *Worker {
	return &Worker{queue: queue, pollInterval: pollInterval, handlers: handlers}
}
func (worker *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(worker.pollInterval)
	defer ticker.Stop()
	for {
		if err := worker.process(ctx); err != nil && ctx.Err() == nil {
			log.Printf("job worker: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
func (worker *Worker) process(ctx context.Context) error {
	job, err := worker.queue.Claim(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	handler, ok := worker.handlers[job.Type]
	if !ok {
		return worker.queue.Fail(ctx, job.ID, "unsupported job type")
	}
	if err := handler(ctx, job); err != nil {
		return worker.queue.Fail(ctx, job.ID, err.Error())
	}
	return worker.queue.Complete(ctx, job.ID)
}
