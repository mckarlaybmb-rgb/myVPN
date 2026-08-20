package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeStore struct {
	linked  bool
	upserts int
}

func (f *fakeStore) Upsert(context.Context, User) error               { f.upserts++; return nil }
func (f *fakeStore) LinkByEmail(context.Context, int64, string) error { f.linked = true; return nil }
func (f *fakeStore) Get(context.Context, int64) (User, error) {
	if f.linked {
		return User{UserID: "user-1"}, nil
	}
	return User{}, nil
}
func (f *fakeStore) RecordNotification(context.Context, int64, string, string) (bool, error) {
	return true, nil
}
func testBot(t *testing.T, store Store) *Bot {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	bot := NewBot("token", store)
	bot.apiBase = server.URL
	return bot
}
func TestBotLinksExistingAccount(t *testing.T) {
	store := &fakeStore{}
	bot := testBot(t, store)
	message := &Message{}
	message.From.ID = 1
	message.Text = "/start user@example.com"
	if err := bot.handle(context.Background(), message); err != nil {
		t.Fatal(err)
	}
	if !store.linked || store.upserts != 1 {
		t.Fatalf("store=%#v", store)
	}
}
func TestBotRequiresLinkBeforeVPNStatus(t *testing.T) {
	store := &fakeStore{}
	bot := testBot(t, store)
	message := &Message{}
	message.From.ID = 1
	message.Text = "/status"
	if err := bot.handle(context.Background(), message); err != nil {
		t.Fatal(err)
	}
}
