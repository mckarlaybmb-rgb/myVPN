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

type fakeProvisioner struct {
	created, deleted string
}

func (fake *fakeProvisioner) CreateClient(_ context.Context, userID, _ string) (models.XrayClient, error) {
	fake.created = userID
	return models.XrayClient{UserID: userID}, nil
}
func (fake *fakeProvisioner) DeleteClient(_ context.Context, userID string) error {
	fake.deleted = userID
	return nil
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

func TestUserServiceProvisionsClient(t *testing.T) {
	repository := &fakeUsers{}
	provisioner := &fakeProvisioner{}
	user, err := NewUserService(repository, provisioner).Create(context.Background(), "user@example.com")
	if err != nil || user.ID != "user-1" || provisioner.created != user.ID {
		t.Fatalf("user=%#v provisioner=%#v err=%v", user, provisioner, err)
	}
	if err := NewUserService(repository, provisioner).Delete(context.Background(), user.ID); err != nil || provisioner.deleted != user.ID {
		t.Fatalf("delete err=%v provisioner=%#v", err, provisioner)
	}
}
