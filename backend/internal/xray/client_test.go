package xray

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

func TestClientAddAndRemoveUser(t *testing.T) {
	inbound := map[string]any{"id": float64(7), "protocol": "vless", "settings": `{"clients":[{"id":"old","email":"old@example.com"}]}`, "streamSettings": "{}"}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/login":
			if err := request.ParseForm(); err != nil || request.Form.Get("username") != "admin" || request.Form.Get("password") != "secret" {
				t.Errorf("login form: %v", request.Form)
			}
			http.SetCookie(writer, &http.Cookie{Name: "session", Value: "ok", Path: "/"})
			_, _ = writer.Write([]byte(`{"success":true,"msg":"success"}`))
		case "/panel/api/inbounds/get/7":
			if _, err := request.Cookie("session"); err != nil {
				t.Error("missing login cookie")
			}
			encoded, _ := json.Marshal(inbound)
			_, _ = writer.Write([]byte(`{"success":true,"obj":` + string(encoded) + `}`))
		case "/panel/api/inbounds/update/7":
			if request.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("content type=%q", request.Header.Get("Content-Type"))
			}
			if err := request.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(request.Form.Get("settings")), &map[string]any{}); err != nil {
				t.Fatalf("settings form: %v", err)
			}
			_, _ = writer.Write([]byte(`{"success":true}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Username: "admin", Password: "secret", InboundID: 7})
	vpnClient := models.XrayClient{UUID: "12345678-1234-4234-8234-123456789abc", Email: "new@example.com", Config: map[string]any{"flow": "xtls-rprx-vision"}}
	subscriptionURL, err := client.AddUser(context.Background(), vpnClient)
	if err != nil || subscriptionURL != server.URL+"/sub/12345678123442348234123456789abc" {
		t.Fatalf("subscription URL=%q err=%v", subscriptionURL, err)
	}
	if err := client.RemoveUser(context.Background(), vpnClient); err != nil {
		t.Fatal(err)
	}
}

func TestClientEnableAndDisableUpdateExistingClient(t *testing.T) {
	inbound := map[string]any{"id": float64(7), "settings": `{"clients":[{"id":"12345678-1234-4234-8234-123456789abc","email":"vpn@example.com","enable":false}]}`}
	updates := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/login":
			_, _ = writer.Write([]byte(`{"success":true}`))
		case "/panel/api/inbounds/get/7":
			encoded, _ := json.Marshal(inbound)
			_, _ = writer.Write([]byte(`{"success":true,"obj":` + string(encoded) + `}`))
		case "/panel/api/inbounds/update/7":
			updates++
			if err := request.ParseForm(); err != nil {
				t.Fatal(err)
			}
			inbound["settings"] = request.Form.Get("settings")
			_, _ = writer.Write([]byte(`{"success":true}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, Username: "admin", Password: "secret", InboundID: 7})
	vpnClient := models.XrayClient{UUID: "12345678-1234-4234-8234-123456789abc", Email: "vpn@example.com"}
	if err := client.EnableClient(context.Background(), vpnClient); err != nil {
		t.Fatal(err)
	}
	if err := client.DisableClient(context.Background(), vpnClient); err != nil {
		t.Fatal(err)
	}
	if updates != 2 {
		t.Fatalf("update calls=%d", updates)
	}
	var settings struct {
		Clients []map[string]any `json:"clients"`
	}
	if err := json.Unmarshal([]byte(inbound["settings"].(string)), &settings); err != nil {
		t.Fatal(err)
	}
	if len(settings.Clients) != 1 || settings.Clients[0]["enable"] != false {
		t.Fatalf("clients=%#v", settings.Clients)
	}
}

func TestClientRejectsInvalidConfiguration(t *testing.T) {
	client := NewClient(Config{BaseURL: "https://x-ui.example"})
	_, err := client.AddUser(context.Background(), models.XrayClient{})
	if err == nil || strings.Contains(err.Error(), url.QueryEscape("secret")) {
		t.Fatalf("unexpected error: %v", err)
	}
}
