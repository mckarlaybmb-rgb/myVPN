package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type AdminRepository struct{ pool *pgxpool.Pool }

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository { return &AdminRepository{pool: pool} }

func (r *AdminRepository) Stats(ctx context.Context) (map[string]int, error) {
	stats := map[string]int{}
	queries := map[string]string{"users": `SELECT COUNT(*) FROM users`, "active_subscriptions": `SELECT COUNT(*) FROM subscriptions WHERE status = 'active' AND expires_at > NOW()`, "expired_subscriptions": `SELECT COUNT(*) FROM subscriptions WHERE status = 'expired' OR expires_at <= NOW()`, "active_vpn_accounts": `SELECT COUNT(*) FROM xray_clients WHERE enabled`, "nodes": `SELECT COUNT(*) FROM vpn_nodes`}
	for name, query := range queries {
		var count int
		if err := r.pool.QueryRow(ctx, query).Scan(&count); err != nil {
			return nil, err
		}
		stats[name] = count
	}
	return stats, nil
}
func (r *AdminRepository) ListNodes(ctx context.Context) ([]models.Node, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text, name, address, port, status, last_check_at, latency_ms FROM vpn_nodes ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Node{}
	for rows.Next() {
		var item models.Node
		if err := rows.Scan(&item.ID, &item.Name, &item.Address, &item.Port, &item.Status, &item.LastCheckAt, &item.LatencyMS); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *AdminRepository) ListSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text, user_id::text, plan, status, expires_at, created_at, updated_at FROM subscriptions ORDER BY expires_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Subscription{}
	for rows.Next() {
		var item models.Subscription
		if err := rows.Scan(&item.ID, &item.UserID, &item.Plan, &item.Status, &item.ExpiresAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *AdminRepository) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text, email, created_at, updated_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.User{}
	for rows.Next() {
		var item models.User
		if err := rows.Scan(&item.ID, &item.Email, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *AdminRepository) UpdateNodeHealth(ctx context.Context, id, status string, latencyMS *int) error {
	_, err := r.pool.Exec(ctx, `UPDATE vpn_nodes SET status = $2, latency_ms = $3, last_check_at = NOW(), updated_at = NOW() WHERE id = $1`, id, status, latencyMS)
	return err
}
func (r *AdminRepository) ListActive(ctx context.Context) ([]models.Node, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text, name, address, port, status, last_check_at, latency_ms FROM vpn_nodes WHERE status <> 'disabled'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Node{}
	for rows.Next() {
		var item models.Node
		if err := rows.Scan(&item.ID, &item.Name, &item.Address, &item.Port, &item.Status, &item.LastCheckAt, &item.LatencyMS); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
