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
type EventNotifier interface {
	NotifyUser(context.Context, string, string, string) error
}

type UserService struct {
	repository  UserRepository
	provisioner ClientProvisioner
	notifier    EventNotifier
}

func NewUserService(repository UserRepository, dependencies ...any) *UserService {
	var provisioner ClientProvisioner
	var notifier EventNotifier
	for _, dependency := range dependencies {
		if value, ok := dependency.(ClientProvisioner); ok {
			provisioner = value
		}
		if value, ok := dependency.(EventNotifier); ok {
			notifier = value
		}
	}
	return &UserService{repository: repository, provisioner: provisioner, notifier: notifier}
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
		if service.notifier != nil {
			_ = service.notifier.NotifyUser(ctx, user.ID, "critical.system_error", user.ID)
		}
		_ = service.repository.Delete(ctx, user.ID)
		return models.User{}, fmt.Errorf("provision xray client: %w", err)
	}
	if service.notifier != nil {
		_ = service.notifier.NotifyUser(ctx, user.ID, "account.created", user.ID)
		_ = service.notifier.NotifyUser(ctx, user.ID, "vpn.provisioned", user.ID)
	}
	return user, nil
}
func (service *UserService) Delete(ctx context.Context, id string) error {
	if service.provisioner != nil {
		if err := service.provisioner.DeleteClient(ctx, id); err != nil {
			return fmt.Errorf("delete xray client: %w", err)
		}
	}
	if service.notifier != nil {
		_ = service.notifier.NotifyUser(ctx, id, "vpn.account_deleted", id)
	}
	if err := service.repository.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}
