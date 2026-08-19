package services

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type UserService struct{ pool *pgxpool.Pool }
func NewUserService(pool *pgxpool.Pool) *UserService { return &UserService{pool: pool} }
func (service *UserService) List(ctx context.Context) ([]models.User, error) {
	rows, err := service.pool.Query(ctx, `SELECT id::text, email, created_at FROM users ORDER BY created_at DESC`); if err != nil { return nil, err }; defer rows.Close()
	users := make([]models.User, 0)
	for rows.Next() { var user models.User; if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil { return nil, err }; users = append(users, user) }
	return users, rows.Err()
}
func (service *UserService) Create(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := service.pool.QueryRow(ctx, `INSERT INTO users (email) VALUES ($1) RETURNING id::text, email, created_at`, email).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil { return user, fmt.Errorf("create user: %w", err) }; return user, nil
}