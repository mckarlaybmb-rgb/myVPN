package repositories

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type XrayClientRepository struct{ pool *pgxpool.Pool }

func NewXrayClientRepository(pool *pgxpool.Pool) *XrayClientRepository {
	return &XrayClientRepository{pool: pool}
}

func (repository *XrayClientRepository) Create(ctx context.Context, client models.XrayClient) (models.XrayClient, error) {
	config, err := json.Marshal(client.Config)
	if err != nil {
		return models.XrayClient{}, err
	}
	return repository.scanClient(repository.pool.QueryRow(ctx, `
		INSERT INTO xray_clients (user_id, email, uuid, subscription_url, protocol, config, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, user_id::text, email, uuid::text, subscription_url, protocol, config, enabled, created_at, updated_at`,
		client.UserID, client.Email, client.UUID, client.SubscriptionURL, client.Protocol, config, client.Enabled))
}

func (repository *XrayClientRepository) GetByUser(ctx context.Context, userID string) (models.XrayClient, error) {
	return repository.scanClient(repository.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, email, uuid::text, subscription_url, protocol, config, enabled, created_at, updated_at
		FROM xray_clients WHERE user_id = $1`, userID))
}

func (repository *XrayClientRepository) Delete(ctx context.Context, userID string) error {
	result, err := repository.pool.Exec(ctx, `DELETE FROM xray_clients WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (repository *XrayClientRepository) SetEnabled(ctx context.Context, userID string, enabled bool) error {
	result, err := repository.pool.Exec(ctx, `UPDATE xray_clients SET enabled = $2, updated_at = NOW() WHERE user_id = $1`, userID, enabled)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type clientRow interface {
	Scan(...any) error
}

func (repository *XrayClientRepository) scanClient(row clientRow) (models.XrayClient, error) {
	var client models.XrayClient
	var config []byte
	if err := row.Scan(&client.ID, &client.UserID, &client.Email, &client.UUID, &client.SubscriptionURL, &client.Protocol, &config, &client.Enabled, &client.CreatedAt, &client.UpdatedAt); err != nil {
		return client, err
	}
	if err := json.Unmarshal(config, &client.Config); err != nil {
		return client, err
	}
	return client, nil
}
