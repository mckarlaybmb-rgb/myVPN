package repositories

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"time"
)

var ErrNotFound = pgx.ErrNoRows

type SubscriptionRepository struct{ pool *pgxpool.Pool }

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}
func (repository *SubscriptionRepository) ListByUser(ctx context.Context, userID string) ([]models.Subscription, error) {
	rows, err := repository.pool.Query(ctx, `SELECT id::text, user_id::text, plan, status, expires_at, created_at FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.Subscription, 0)
	for rows.Next() {
		var item models.Subscription
		if err := rows.Scan(&item.ID, &item.UserID, &item.Plan, &item.Status, &item.ExpiresAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (repository *SubscriptionRepository) Create(ctx context.Context, userID, plan string, expiresAt time.Time) (models.Subscription, error) {
	var item models.Subscription
	err := repository.pool.QueryRow(ctx, `INSERT INTO subscriptions (user_id, plan, expires_at) VALUES ($1, $2, $3) RETURNING id::text, user_id::text, plan, status, expires_at, created_at`, userID, plan, expiresAt).Scan(&item.ID, &item.UserID, &item.Plan, &item.Status, &item.ExpiresAt, &item.CreatedAt)
	return item, err
}
func (repository *SubscriptionRepository) Renew(ctx context.Context, id string, extraDays int) (models.Subscription, error) {
	var item models.Subscription
	err := repository.pool.QueryRow(ctx, `UPDATE subscriptions SET expires_at = expires_at + ($2 * INTERVAL '1 day'), status = 'active' WHERE id = $1 RETURNING id::text, user_id::text, plan, status, expires_at, created_at`, id, extraDays).Scan(&item.ID, &item.UserID, &item.Plan, &item.Status, &item.ExpiresAt, &item.CreatedAt)
	return item, err
}

func (repository *SubscriptionRepository) UpdateStatus(ctx context.Context, id, status string) (models.Subscription, error) {
	var item models.Subscription
	err := repository.pool.QueryRow(ctx, `UPDATE subscriptions SET status = $2 WHERE id = $1 RETURNING id::text, user_id::text, plan, status, expires_at, created_at`, id, status).Scan(&item.ID, &item.UserID, &item.Plan, &item.Status, &item.ExpiresAt, &item.CreatedAt)
	return item, err
}
