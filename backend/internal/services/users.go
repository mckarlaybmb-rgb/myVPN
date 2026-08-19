package services

import (
	"context"
	"fmt"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type UserRepository interface {
	List(context.Context) ([]models.User, error)
	Create(context.Context, string) (models.User, error)
	Delete(context.Context, string) error
}
type ClientProvisioner interface {
	CreateClient(context.Context, string, string) (models.XrayClient, error)
	DeleteClient(context.Context, string) error
}

type UserService struct {
	repository  UserRepository
	provisioner ClientProvisioner
}

func NewUserService(repository UserRepository, provisioners ...ClientProvisioner) *UserService {
	var provisioner ClientProvisioner
	if len(provisioners) > 0 {
		provisioner = provisioners[0]
	}
	return &UserService{repository: repository, provisioner: provisioner}
}
func (service *UserService) List(ctx context.Context) ([]models.User, error) {
	return service.repository.List(ctx)
}
func (service *UserService) Create(ctx context.Context, email string) (models.User, error) {
	user, err := service.repository.Create(ctx, email)
	if err != nil || service.provisioner == nil {
		return user, err
	}
	if _, err := service.provisioner.CreateClient(ctx, user.ID, user.Email); err != nil {
		_ = service.repository.Delete(ctx, user.ID)
		return models.User{}, fmt.Errorf("provision xray client: %w", err)
	}
	return user, nil
}
func (service *UserService) Delete(ctx context.Context, id string) error {
	if service.provisioner != nil {
		if err := service.provisioner.DeleteClient(ctx, id); err != nil {
			return fmt.Errorf("delete xray client: %w", err)
		}
	}
	return service.repository.Delete(ctx, id)
}
