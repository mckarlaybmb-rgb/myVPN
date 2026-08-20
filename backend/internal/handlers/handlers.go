package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	users         *services.UserService
	subscriptions *services.SubscriptionService
	admin         AdminReader
}

type AdminReader interface {
	Stats(context.Context) (map[string]int, error)
	ListNodes(context.Context) ([]models.Node, error)
	ListSubscriptions(context.Context) ([]models.Subscription, error)
	ListUsers(context.Context) ([]models.User, error)
}

func New(users *services.UserService, subscriptions *services.SubscriptionService, admin ...AdminReader) *Handler {
	var reader AdminReader
	if len(admin) > 0 {
		reader = admin[0]
	}
	return &Handler{users: users, subscriptions: subscriptions, admin: reader}
}

func (handler *Handler) AdminStats(writer http.ResponseWriter, request *http.Request) {
	if handler.admin == nil {
		http.Error(writer, "admin service unavailable", http.StatusServiceUnavailable)
		return
	}
	stats, err := handler.admin.Stats(request.Context())
	writeResult(writer, stats, err, http.StatusOK)
}
func (handler *Handler) AdminNodes(writer http.ResponseWriter, request *http.Request) {
	if handler.admin == nil {
		http.Error(writer, "admin service unavailable", http.StatusServiceUnavailable)
		return
	}
	items, err := handler.admin.ListNodes(request.Context())
	writeResult(writer, items, err, http.StatusOK)
}
func (handler *Handler) AdminSubscriptions(writer http.ResponseWriter, request *http.Request) {
	if handler.admin == nil {
		http.Error(writer, "admin service unavailable", http.StatusServiceUnavailable)
		return
	}
	items, err := handler.admin.ListSubscriptions(request.Context())
	writeResult(writer, items, err, http.StatusOK)
}
func (handler *Handler) AdminUsers(writer http.ResponseWriter, request *http.Request) {
	if handler.admin == nil {
		http.Error(writer, "admin service unavailable", http.StatusServiceUnavailable)
		return
	}
	items, err := handler.admin.ListUsers(request.Context())
	writeResult(writer, items, err, http.StatusOK)
}
func (handler *Handler) Health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}
func (handler *Handler) CreateUser(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if !decode(writer, request, &input) {
		return
	}
	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" {
		http.Error(writer, "email is required", http.StatusBadRequest)
		return
	}
	user, err := handler.users.Create(request.Context(), input.Email)
	writeResult(writer, user, err, http.StatusCreated)
}
func (handler *Handler) ListUsers(writer http.ResponseWriter, request *http.Request) {
	users, err := handler.users.List(request.Context())
	writeResult(writer, users, err, http.StatusOK)
}
func (handler *Handler) DeleteUser(writer http.ResponseWriter, request *http.Request) {
	err := handler.users.Delete(request.Context(), request.PathValue("id"))
	if err != nil {
		writeError(writer, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
func (handler *Handler) CreateSubscription(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		UserID    string    `json:"user_id"`
		Plan      string    `json:"plan"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if !decode(writer, request, &input) {
		return
	}
	if input.UserID == "" || input.Plan == "" || input.ExpiresAt.IsZero() {
		http.Error(writer, "user_id, plan, and expires_at are required", http.StatusBadRequest)
		return
	}
	item, err := handler.subscriptions.Create(request.Context(), input.UserID, input.Plan, input.ExpiresAt)
	writeResult(writer, item, err, http.StatusCreated)
}
func (handler *Handler) ListSubscriptions(writer http.ResponseWriter, request *http.Request) {
	items, err := handler.subscriptions.ListByUser(request.Context(), request.PathValue("user_id"))
	writeResult(writer, items, err, http.StatusOK)
}
func (handler *Handler) RenewSubscription(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		ExtraDays int `json:"extra_days"`
	}
	if !decode(writer, request, &input) {
		return
	}
	item, err := handler.subscriptions.Renew(request.Context(), request.PathValue("id"), input.ExtraDays)
	writeResult(writer, item, err, http.StatusOK)
}
func (handler *Handler) SuspendSubscription(writer http.ResponseWriter, request *http.Request) {
	item, err := handler.subscriptions.Suspend(request.Context(), request.PathValue("id"))
	writeResult(writer, item, err, http.StatusOK)
}
func (handler *Handler) ExpireSubscription(writer http.ResponseWriter, request *http.Request) {
	item, err := handler.subscriptions.Expire(request.Context(), request.PathValue("id"))
	writeResult(writer, item, err, http.StatusOK)
}
func decode(writer http.ResponseWriter, request *http.Request, target any) bool {
	if err := json.NewDecoder(request.Body).Decode(target); err != nil {
		http.Error(writer, "invalid JSON body", http.StatusBadRequest)
		return false
	}
	return true
}
func writeResult(writer http.ResponseWriter, value any, err error, successStatus int) {
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, successStatus, value)
}
func writeError(writer http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, pgx.ErrNoRows) {
		status = http.StatusNotFound
	}
	http.Error(writer, err.Error(), status)
}
func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
