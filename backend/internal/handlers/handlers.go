package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/services"
)

type Handler struct { users *services.UserService; subscriptions *services.SubscriptionService }
func New(users *services.UserService, subscriptions *services.SubscriptionService) *Handler { return &Handler{users: users, subscriptions: subscriptions} }
func (handler *Handler) Health(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ok"}) }
func (handler *Handler) Users(w http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet { users, err := handler.users.List(request.Context()); writeResult(w, users, err); return }
	if request.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	var input struct { Email string `json:"email"` }; if !decode(w, request, &input) || strings.TrimSpace(input.Email) == "" { http.Error(w, "email is required", http.StatusBadRequest); return }
	user, err := handler.users.Create(request.Context(), strings.TrimSpace(input.Email)); writeResultStatus(w, user, err, http.StatusCreated)
}
func (handler *Handler) Subscriptions(w http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet { items, err := handler.subscriptions.List(request.Context()); writeResult(w, items, err); return }
	if request.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	var input struct { UserID string `json:"user_id"`; Plan string `json:"plan"`; ExpiresAt time.Time `json:"expires_at"` }
	if !decode(w, request, &input) { return }; if input.UserID == "" || input.Plan == "" || input.ExpiresAt.IsZero() { http.Error(w, "user_id, plan, and expires_at are required", http.StatusBadRequest); return }
	item, err := handler.subscriptions.Create(request.Context(), input.UserID, input.Plan, input.ExpiresAt); writeResultStatus(w, item, err, http.StatusCreated)
}
func decode(w http.ResponseWriter, request *http.Request, target any) bool { if err := json.NewDecoder(request.Body).Decode(target); err != nil { http.Error(w, "invalid JSON body", http.StatusBadRequest); return false }; return true }
func writeResult(w http.ResponseWriter, value any, err error) { writeResultStatus(w, value, err, http.StatusOK) }
func writeResultStatus(w http.ResponseWriter, value any, err error, successStatus int) { if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; writeJSON(w, successStatus, value) }
func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value) }