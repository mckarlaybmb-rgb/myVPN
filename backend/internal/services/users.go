package services

import (
	"context"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type UserRepository interface {
	List(context.Context) ([]models.User, error)
	Create(context.Context, string) (models.User, error)
	Delete(context.Context, string) error
}
type UserService struct{ repository UserRepository }

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}
func (service *UserService) List(ctx context.Context) ([]models.User, error) {
	return service.repository.List(ctx)
}
func (service *UserService) Create(ctx context.Context, email string) (models.User, error) {
	return service.repository.Create(ctx, email)
}
func (service *UserService) Delete(ctx context.Context, id string) error {
	return service.repository.Delete(ctx, id)
}
