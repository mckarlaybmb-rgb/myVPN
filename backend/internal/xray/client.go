package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type Runtime interface {
	AddUser(context.Context, models.XrayClient) (string, error)
	RemoveUser(context.Context, models.XrayClient) error
	EnableClient(context.Context, models.XrayClient) error
	DisableClient(context.Context, models.XrayClient) error
}

type Config struct {
	BaseURL   string
	Username  string
	Password  string
	InboundID int64
}
type Client struct {
	config     Config
	httpClient *http.Client
}

func NewClient(config Config) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{config: config, httpClient: &http.Client{Jar: jar}}
}

type apiResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"msg"`
	Object  json.RawMessage `json:"obj"`
}

func (client *Client) request(ctx context.Context, method, path string, body io.Reader, contentType string, response any) error {
	request, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(client.config.BaseURL, "/")+path, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	result, err := client.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return fmt.Errorf("x-ui returned HTTP %d", result.StatusCode)
	}
	if err := json.NewDecoder(result.Body).Decode(response); err != nil {
		return fmt.Errorf("decode x-ui response: %w", err)
	}
	return nil
}
func (client *Client) login(ctx context.Context) error {
	if client.config.BaseURL == "" || client.config.Username == "" || client.config.Password == "" || client.config.InboundID <= 0 {
		return fmt.Errorf("x-ui base URL, username, password, and positive inbound ID are required")
	}
	form := url.Values{"username": {client.config.Username}, "password": {client.config.Password}}
	var response apiResponse
	if err := client.request(ctx, http.MethodPost, "/login", strings.NewReader(form.Encode()), "application/x-www-form-urlencoded", &response); err != nil {
		return fmt.Errorf("x-ui login request: %w", err)
	}
	if !response.Success {
		return fmt.Errorf("x-ui login failed: %s", response.Message)
	}
	return nil
}
func (client *Client) inbound(ctx context.Context) (map[string]any, error) {
	var response apiResponse
	path := "/panel/api/inbounds/get/" + strconv.FormatInt(client.config.InboundID, 10)
	if err := client.request(ctx, http.MethodGet, path, strings.NewReader(""), "", &response); err != nil {
		return nil, fmt.Errorf("get x-ui inbound: %w", err)
	}
	if !response.Success {
		return nil, fmt.Errorf("get x-ui inbound failed: %s", response.Message)
	}
	var inbound map[string]any
	if err := json.Unmarshal(response.Object, &inbound); err != nil {
		return nil, fmt.Errorf("decode x-ui inbound: %w", err)
	}
	return inbound, nil
}
func (client *Client) updateInbound(ctx context.Context, inbound map[string]any) error {
	form := url.Values{}
	for key, value := range inbound {
		if key == "id" {
			continue
		}
		form.Set(key, formValue(value))
	}
	var response apiResponse
	path := "/panel/api/inbounds/update/" + strconv.FormatInt(client.config.InboundID, 10)
	if err := client.request(ctx, http.MethodPost, path, strings.NewReader(form.Encode()), "application/x-www-form-urlencoded", &response); err != nil {
		return fmt.Errorf("update x-ui inbound: %w", err)
	}
	if !response.Success {
		return fmt.Errorf("update x-ui inbound failed: %s", response.Message)
	}
	return nil
}

func formValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	if _, ok := value.(map[string]any); ok {
		encoded, _ := json.Marshal(value)
		return string(encoded)
	}
	if _, ok := value.([]any); ok {
		encoded, _ := json.Marshal(value)
		return string(encoded)
	}
	return fmt.Sprint(value)
}

func inboundSettings(inbound map[string]any) (map[string]any, error) {
	settings, ok := inbound["settings"].(string)
	if !ok {
		return nil, fmt.Errorf("x-ui inbound settings are not a JSON string")
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(settings), &result); err != nil {
		return nil, fmt.Errorf("decode x-ui inbound settings: %w", err)
	}
	return result, nil
}
func (client *Client) AddUser(ctx context.Context, vpnClient models.XrayClient) (string, error) {
	if err := client.login(ctx); err != nil {
		return "", err
	}
	inbound, err := client.inbound(ctx)
	if err != nil {
		return "", err
	}
	settings, err := inboundSettings(inbound)
	if err != nil {
		return "", err
	}
	clients, _ := settings["clients"].([]any)
	subID := strings.ReplaceAll(vpnClient.UUID, "-", "")
	for _, entry := range clients {
		item, ok := entry.(map[string]any)
		if ok && (item["id"] == vpnClient.UUID || item["email"] == vpnClient.Email) {
			return strings.TrimRight(client.config.BaseURL, "/") + "/sub/" + subID, nil
		}
	}
	clients = append(clients, map[string]any{"id": vpnClient.UUID, "email": vpnClient.Email, "enable": true, "flow": configString(vpnClient.Config, "flow"), "subId": subID})
	settings["clients"] = clients
	encoded, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	inbound["settings"] = string(encoded)
	if err := client.updateInbound(ctx, inbound); err != nil {
		return "", err
	}
	log.Printf("x-ui client added: email=%s inbound=%d", vpnClient.Email, client.config.InboundID)
	return strings.TrimRight(client.config.BaseURL, "/") + "/sub/" + subID, nil
}
func (client *Client) RemoveUser(ctx context.Context, vpnClient models.XrayClient) error {
	if err := client.login(ctx); err != nil {
		return err
	}
	inbound, err := client.inbound(ctx)
	if err != nil {
		return err
	}
	settings, err := inboundSettings(inbound)
	if err != nil {
		return err
	}
	clients, _ := settings["clients"].([]any)
	filtered := make([]any, 0, len(clients))
	for _, entry := range clients {
		item, ok := entry.(map[string]any)
		if !ok || (item["id"] != vpnClient.UUID && item["email"] != vpnClient.Email) {
			filtered = append(filtered, entry)
		}
	}
	settings["clients"] = filtered
	encoded, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound["settings"] = string(encoded)
	if err := client.updateInbound(ctx, inbound); err != nil {
		return err
	}
	log.Printf("x-ui client removed: email=%s inbound=%d", vpnClient.Email, client.config.InboundID)
	return nil
}
func (client *Client) EnableClient(ctx context.Context, vpnClient models.XrayClient) error {
	return client.setClientEnabled(ctx, vpnClient, true)
}
func (client *Client) DisableClient(ctx context.Context, vpnClient models.XrayClient) error {
	return client.setClientEnabled(ctx, vpnClient, false)
}

func (client *Client) setClientEnabled(ctx context.Context, vpnClient models.XrayClient, enabled bool) error {
	if err := client.login(ctx); err != nil {
		return err
	}
	inbound, err := client.inbound(ctx)
	if err != nil {
		return err
	}
	settings, err := inboundSettings(inbound)
	if err != nil {
		return err
	}
	clients, _ := settings["clients"].([]any)
	found := false
	for _, entry := range clients {
		item, ok := entry.(map[string]any)
		if ok && (item["id"] == vpnClient.UUID || item["email"] == vpnClient.Email) {
			item["enable"] = enabled
			found = true
		}
	}
	if !found {
		return fmt.Errorf("x-ui client not found: %s", vpnClient.Email)
	}
	settings["clients"] = clients
	encoded, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	inbound["settings"] = string(encoded)
	return client.updateInbound(ctx, inbound)
}
func configString(config map[string]any, key string) string {
	value, _ := config[key].(string)
	return value
}
func BuildVLESSConfig(inboundTag string) map[string]any {
	return map[string]any{"protocol": "vless", "inbound_tag": inboundTag, "flow": "xtls-rprx-vision", "encryption": "none"}
}
func ConfigJSON(config map[string]any) ([]byte, error) { return json.Marshal(config) }
