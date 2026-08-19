package services

import (
	"context"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"testing"
	"time"
)

type fakeSubscriptions struct{ renewedDays int }

func (fake *fakeSubscriptions) ListByUser(context.Context, string) ([]models.Subscription, error) {
	return []models.Subscription{}, nil
}
func (fake *fakeSubscriptions) Create(_ context.Context, userID, plan string, expiresAt time.Time) (models.Subscription, error) {
	return models.Subscription{ID: "sub-1", UserID: userID, Plan: plan, ExpiresAt: expiresAt}, nil
}
func (fake *fakeSubscriptions) Renew(_ context.Context, id string, days int) (models.Subscription, error) {
	fake.renewedDays = days
	return models.Subscription{ID: id, UserID: "user-1"}, nil
}

type fakeQueue struct{ types []string }

func (fake *fakeQueue) Enqueue(_ context.Context, jobType, _ string, _ map[string]string) (string, error) {
	fake.types = append(fake.types, jobType)
	return "job-1", nil
}
func TestSubscriptionServiceEnqueuesCreateAndRenew(t *testing.T) {
	repository := &fakeSubscriptions{}
	queue := &fakeQueue{}
	service := NewSubscriptionService(repository, queue)
	ctx := context.Background()
	if _, err := service.Create(ctx, "user-1", "monthly", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Renew(ctx, "sub-1", 30); err != nil {
		t.Fatal(err)
	}
	if repository.renewedDays != 30 || len(queue.types) != 2 || queue.types[0] != "subscription.created" || queue.types[1] != "subscription.renewed" {
		t.Fatalf("repository=%#v queue=%#v", repository, queue)
	}
	if _, err := service.Renew(ctx, "sub-1", 0); err == nil {
		t.Fatal("expected invalid renewal error")
	}
}
