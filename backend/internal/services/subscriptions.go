package services

import (
	"context"
	"fmt"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/jobs"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"time"
)

type SubscriptionRepository interface {
	ListByUser(context.Context, string) ([]models.Subscription, error)
	Create(context.Context, string, string, time.Time) (models.Subscription, error)
	Renew(context.Context, string, int) (models.Subscription, error)
}
type JobEnqueuer interface {
	Enqueue(context.Context, string, string, map[string]string) (string, error)
}
type SubscriptionService struct {
	repository SubscriptionRepository
	clients    ClientDisabler
	enabler    ClientEnabler
	notifier   EventNotifier
}

type ClientDisabler interface {
	DisableClient(context.Context, string) error
}
type ClientEnabler interface {
	EnableClient(context.Context, string) error
}

type SubscriptionStatusRepository interface {
	UpdateStatus(context.Context, string, string) (models.Subscription, error)
}

func NewSubscriptionService(repository SubscriptionRepository, _ JobEnqueuer, dependencies ...any) *SubscriptionService {
	var clientDisabler ClientDisabler
	var enabler ClientEnabler
	var notifier EventNotifier
	for _, dependency := range dependencies {
		if value, ok := dependency.(ClientDisabler); ok {
			clientDisabler = value
		}
		if value, ok := dependency.(ClientEnabler); ok {
			enabler = value
		}
		if value, ok := dependency.(EventNotifier); ok {
			notifier = value
		}
	}
	return &SubscriptionService{repository: repository, clients: clientDisabler, enabler: enabler, notifier: notifier}
}
func (service *SubscriptionService) ListByUser(ctx context.Context, userID string) ([]models.Subscription, error) {
	return service.repository.ListByUser(ctx, userID)
}
func (service *SubscriptionService) Create(ctx context.Context, userID, plan string, expiresAt time.Time) (models.Subscription, error) {
	item, err := service.repository.Create(ctx, userID, plan, expiresAt)
	if err != nil {
		return item, fmt.Errorf("create subscription: %w", err)
	}
	if service.notifier != nil {
		_ = service.notifier.NotifyUser(ctx, userID, "subscription.created", item.ID)
	}
	return item, nil
}
func (service *SubscriptionService) Renew(ctx context.Context, id string, extraDays int) (models.Subscription, error) {
	if extraDays <= 0 {
		return models.Subscription{}, fmt.Errorf("extra_days must be greater than zero")
	}
	item, err := service.repository.Renew(ctx, id, extraDays)
	if err != nil {
		return item, fmt.Errorf("renew subscription: %w", err)
	}
	if service.enabler != nil {
		if err := service.enabler.EnableClient(ctx, item.UserID); err != nil {
			return item, fmt.Errorf("enable client after renewal: %w", err)
		}
		if service.notifier != nil {
			_ = service.notifier.NotifyUser(ctx, item.UserID, "vpn.account_re_enabled", item.ID)
		}
	}
	if service.notifier != nil {
		_ = service.notifier.NotifyUser(ctx, item.UserID, "subscription.renewed", item.ID)
	}
	return item, nil
}

func (service *SubscriptionService) Suspend(ctx context.Context, id string) (models.Subscription, error) {
	return service.updateStatus(ctx, id, "suspended")
}

func (service *SubscriptionService) Expire(ctx context.Context, id string) (models.Subscription, error) {
	return service.updateStatus(ctx, id, "expired")
}

func (service *SubscriptionService) updateStatus(ctx context.Context, id, status string) (models.Subscription, error) {
	repository, ok := service.repository.(SubscriptionStatusRepository)
	if !ok {
		return models.Subscription{}, fmt.Errorf("subscription status updates are not supported")
	}
	item, err := repository.UpdateStatus(ctx, id, status)
	if err != nil {
		return item, fmt.Errorf("update subscription status: %w", err)
	}
	if service.clients != nil {
		if err := service.clients.DisableClient(ctx, item.UserID); err != nil {
			return item, fmt.Errorf("disable client for %s subscription: %w", status, err)
		}
	}
	if service.notifier != nil {
		event := "vpn.account_suspended"
		if status == "expired" {
			event = "subscription.expired"
		}
		_ = service.notifier.NotifyUser(ctx, item.UserID, event, item.ID)
	}
	return item, nil
}

var _ JobEnqueuer = (*jobs.Queue)(nil)
