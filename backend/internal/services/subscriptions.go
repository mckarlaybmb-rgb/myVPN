package services

import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type SubscriptionService struct { pool *pgxpool.Pool; queue *jobs.Queue }
func NewSubscriptionService(pool *pgxpool.Pool, queue *jobs.Queue) *SubscriptionService { return &SubscriptionService{pool: pool, queue: queue} }
func (service *SubscriptionService) List(ctx context.Context) ([]models.Subscription, error) {
	rows, err := service.pool.Query(ctx, `SELECT id::text, user_id::text, plan, status, expires_at, created_at FROM subscriptions ORDER BY created_at DESC`); if err != nil { return nil, err }; defer rows.Close()
	items := make([]models.Subscription, 0)
	for rows.Next() { var item models.Subscription; if err := rows.Scan(&item.ID, &item.UserID, &item.Plan, &item.Status, &item.ExpiresAt, &item.CreatedAt); err != nil { return nil, err }; items = append(items, item) }
	return items, rows.Err()
}
func (service *SubscriptionService) Create(ctx context.Context, userID, plan string, expiresAt time.Time) (models.Subscription, error) {
	var item models.Subscription
	err := service.pool.QueryRow(ctx, `INSERT INTO subscriptions (user_id, plan, expires_at) VALUES ($1, $2, $3) RETURNING id::text, user_id::text, plan, status, expires_at, created_at`, userID, plan, expiresAt).Scan(&item.ID, &item.UserID, &item.Plan, &item.Status, &item.ExpiresAt, &item.CreatedAt)
	if err != nil { return item, fmt.Errorf("create subscription: %w", err) }
	if _, err := service.queue.Enqueue(ctx, "subscription.created", item.ID, map[string]string{"user_id": userID}); err != nil { return item, fmt.Errorf("enqueue subscription job: %w", err) }
	return item, nil
}