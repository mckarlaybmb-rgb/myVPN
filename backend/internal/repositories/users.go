package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type UserRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) *UserRepository { return &UserRepository{pool: pool} }
func (repository *UserRepository) List(ctx context.Context) ([]models.User, error) {
	rows, err := repository.pool.Query(ctx, `SELECT id::text, email, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
func (repository *UserRepository) Create(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := repository.pool.QueryRow(ctx, `INSERT INTO users (email) VALUES ($1) RETURNING id::text, email, created_at`, email).Scan(&user.ID, &user.Email, &user.CreatedAt)
	return user, err
}
func (repository *UserRepository) Delete(ctx context.Context, id string) error {
	result, err := repository.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
