package services

import (
	"context"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"testing"
	"time"
)

type fakeUsers struct {
	users   []models.User
	deleted string
}

func (fake *fakeUsers) List(context.Context) ([]models.User, error) { return fake.users, nil }
func (fake *fakeUsers) Create(_ context.Context, email string) (models.User, error) {
	user := models.User{ID: "user-1", Email: email, CreatedAt: time.Now()}
	fake.users = append(fake.users, user)
	return user, nil
}
func (fake *fakeUsers) Delete(_ context.Context, id string) error { fake.deleted = id; return nil }
func TestUserServiceDelegatesOperations(t *testing.T) {
	fake := &fakeUsers{}
	service := NewUserService(fake)
	ctx := context.Background()
	user, err := service.Create(ctx, "user@example.com")
	if err != nil || user.Email != "user@example.com" {
		t.Fatalf("create user: %#v, %v", user, err)
	}
	if _, err := service.List(ctx); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(ctx, "user-1"); err != nil {
		t.Fatal(err)
	}
	if fake.deleted != "user-1" {
		t.Fatalf("deleted %q", fake.deleted)
	}
}
