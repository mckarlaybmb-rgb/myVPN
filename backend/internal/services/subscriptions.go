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
	queue      JobEnqueuer
}

func NewSubscriptionService(repository SubscriptionRepository, queue JobEnqueuer) *SubscriptionService {
	return &SubscriptionService{repository: repository, queue: queue}
}
func (service *SubscriptionService) ListByUser(ctx context.Context, userID string) ([]models.Subscription, error) {
	return service.repository.ListByUser(ctx, userID)
}
func (service *SubscriptionService) Create(ctx context.Context, userID, plan string, expiresAt time.Time) (models.Subscription, error) {
	item, err := service.repository.Create(ctx, userID, plan, expiresAt)
	if err != nil {
		return item, fmt.Errorf("create subscription: %w", err)
	}
	if _, err := service.queue.Enqueue(ctx, "subscription.created", item.ID, map[string]string{"user_id": userID}); err != nil {
		return item, fmt.Errorf("enqueue subscription job: %w", err)
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
	if _, err := service.queue.Enqueue(ctx, "subscription.renewed", item.ID, map[string]string{"user_id": item.UserID}); err != nil {
		return item, fmt.Errorf("enqueue renewal job: %w", err)
	}
	return item, nil
}

var _ JobEnqueuer = (*jobs.Queue)(nil)
