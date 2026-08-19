package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Status string

const (
	Pending Status = "pending"
	Processing Status = "processing"
	Completed Status = "completed"
	Failed Status = "failed"
)

type Job struct { ID string; Type string; EntityID string; Payload []byte }

type Queue struct{ pool *pgxpool.Pool }
func NewQueue(pool *pgxpool.Pool) *Queue { return &Queue{pool: pool} }
func (queue *Queue) Enqueue(ctx context.Context, jobType, entityID string, payload map[string]string) (string, error) {
	data, err := json.Marshal(payload); if err != nil { return "", err }
	var id string
	err = queue.pool.QueryRow(ctx, `INSERT INTO job_queue (job_type, entity_id, payload) VALUES ($1, $2, $3) RETURNING id::text`, jobType, entityID, data).Scan(&id)
	if err != nil { return "", fmt.Errorf("enqueue job: %w", err) }
	return id, nil
}

func (queue *Queue) Claim(ctx context.Context) (Job, error) {
	transaction, err := queue.pool.Begin(ctx)
	if err != nil { return Job{}, err }
	defer transaction.Rollback(ctx)
	var job Job
	err = transaction.QueryRow(ctx, `WITH next_job AS (SELECT id FROM job_queue WHERE status = 'pending' AND available_at <= NOW() ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1) UPDATE job_queue SET status = 'processing', attempts = attempts + 1 WHERE id = (SELECT id FROM next_job) RETURNING id::text, job_type, COALESCE(entity_id::text, ''), payload`).Scan(&job.ID, &job.Type, &job.EntityID, &job.Payload)
	if err != nil { return Job{}, err }
	if err := transaction.Commit(ctx); err != nil { return Job{}, err }
	return job, nil
}

func (queue *Queue) Complete(ctx context.Context, id string) error {
	_, err := queue.pool.Exec(ctx, `UPDATE job_queue SET status = 'completed', processed_at = NOW() WHERE id = $1 AND status = 'processing'`, id)
	return err
}

func (queue *Queue) Fail(ctx context.Context, id, message string) error {
	_, err := queue.pool.Exec(ctx, `UPDATE job_queue SET status = 'failed', last_error = $2, processed_at = NOW() WHERE id = $1 AND status = 'processing'`, id, message)
	return err
}