package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
)

type fakeUserRepository struct {
	users   []models.User
	created string
	deleted string
}

func (repository *fakeUserRepository) List(context.Context) ([]models.User, error) {
	return repository.users, nil
}

func (repository *fakeUserRepository) Create(_ context.Context, email string) (models.User, error) {
	repository.created = email
	user := models.User{ID: "user-1", Email: email, CreatedAt: time.Now()}
	repository.users = append(repository.users, user)
	return user, nil
}

func (repository *fakeUserRepository) Delete(_ context.Context, id string) error {
	repository.deleted = id
	return nil
}

type fakeSubscriptionRepository struct {
	createdUserID string
	createdPlan   string
	renewedID     string
	renewedDays   int
}

func (repository *fakeSubscriptionRepository) ListByUser(_ context.Context, userID string) ([]models.Subscription, error) {
	return []models.Subscription{{ID: "sub-1", UserID: userID, Plan: "monthly"}}, nil
}

func (repository *fakeSubscriptionRepository) Create(_ context.Context, userID, plan string, expiresAt time.Time) (models.Subscription, error) {
	repository.createdUserID = userID
	repository.createdPlan = plan
	return models.Subscription{ID: "sub-1", UserID: userID, Plan: plan, ExpiresAt: expiresAt}, nil
}

func (repository *fakeSubscriptionRepository) Renew(_ context.Context, id string, extraDays int) (models.Subscription, error) {
	repository.renewedID = id
	repository.renewedDays = extraDays
	return models.Subscription{ID: id, UserID: "user-1", Plan: "monthly"}, nil
}

type fakeQueue struct{ jobTypes []string }

func (queue *fakeQueue) Enqueue(_ context.Context, jobType, _ string, _ map[string]string) (string, error) {
	queue.jobTypes = append(queue.jobTypes, jobType)
	return "job-1", nil
}

func newTestHandler() (*Handler, *fakeUserRepository, *fakeSubscriptionRepository, *fakeQueue) {
	users := &fakeUserRepository{}
	subscriptions := &fakeSubscriptionRepository{}
	queue := &fakeQueue{}
	return New(services.NewUserService(users), services.NewSubscriptionService(subscriptions, queue)), users, subscriptions, queue
}

func TestHealth(t *testing.T) {
	handler, _, _, _ := newTestHandler()
	response := httptest.NewRecorder()
	handler.Health(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusOK || response.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected health response: %d %q", response.Code, response.Body.String())
	}
}

func TestUserHandlers(t *testing.T) {
	handler, users, _, _ := newTestHandler()
	createResponse := httptest.NewRecorder()
	handler.CreateUser(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"email":" user@example.com "}`)))
	if createResponse.Code != http.StatusCreated || users.created != "user@example.com" {
		t.Fatalf("unexpected create response: %d, created=%q", createResponse.Code, users.created)
	}

	listResponse := httptest.NewRecorder()
	handler.ListUsers(listResponse, httptest.NewRequest(http.MethodGet, "/api/v1/users", nil))
	if listResponse.Code != http.StatusOK {
		t.Fatalf("unexpected list status: %d", listResponse.Code)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/users/user-1", nil)
	deleteRequest.SetPathValue("id", "user-1")
	deleteResponse := httptest.NewRecorder()
	handler.DeleteUser(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent || users.deleted != "user-1" {
		t.Fatalf("unexpected delete response: %d, deleted=%q", deleteResponse.Code, users.deleted)
	}
}

func TestSubscriptionHandlers(t *testing.T) {
	handler, _, subscriptions, queue := newTestHandler()
	expiresAt := "2030-01-01T00:00:00Z"
	createResponse := httptest.NewRecorder()
	handler.CreateSubscription(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(`{"user_id":"user-1","plan":"monthly","expires_at":"`+expiresAt+`"}`)))
	if createResponse.Code != http.StatusCreated || subscriptions.createdUserID != "user-1" || len(queue.jobTypes) != 0 {
		t.Fatalf("unexpected create response: %d, repository=%#v queue=%#v", createResponse.Code, subscriptions, queue)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/user-1", nil)
	listRequest.SetPathValue("user_id", "user-1")
	listResponse := httptest.NewRecorder()
	handler.ListSubscriptions(listResponse, listRequest)
	var items []models.Subscription
	if listResponse.Code != http.StatusOK || json.NewDecoder(listResponse.Body).Decode(&items) != nil || len(items) != 1 {
		t.Fatalf("unexpected list response: %d %q", listResponse.Code, listResponse.Body.String())
	}

	renewRequest := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-1/renew", strings.NewReader(`{"extra_days":30}`))
	renewRequest.SetPathValue("id", "sub-1")
	renewResponse := httptest.NewRecorder()
	handler.RenewSubscription(renewResponse, renewRequest)
	if renewResponse.Code != http.StatusOK || subscriptions.renewedID != "sub-1" || subscriptions.renewedDays != 30 || len(queue.jobTypes) != 0 {
		t.Fatalf("unexpected renew response: %d, repository=%#v queue=%#v", renewResponse.Code, subscriptions, queue)
	}
}

func TestCreateUserRejectsMissingEmail(t *testing.T) {
	handler, _, _, _ := newTestHandler()
	response := httptest.NewRecorder()
	handler.CreateUser(response, httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"email":" "}`)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}
