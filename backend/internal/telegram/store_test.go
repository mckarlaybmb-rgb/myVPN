package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeStore struct {
	linked       bool
	upserts      int
	records      int
	sent         bool
	recordResult bool
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
	f.records++
	f.sent = true
	return f.recordResult, nil
}
func (f *fakeStore) IsNotificationSent(context.Context, int64, string, string) (bool, error) {
	return f.sent, nil
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

func TestNotifierRecordsSuccessfulDeliveryAfterSending(t *testing.T) {
	store := &fakeStore{recordResult: true}
	bot := testBot(t, store)
	notifier := &Notifier{bot: bot}
	if err := notifier.deliver(context.Background(), 1, "subscription.created", "sub-1"); err != nil {
		t.Fatal(err)
	}
	if store.records != 1 || !store.sent {
		t.Fatalf("records=%d sent=%v", store.records, store.sent)
	}
}

func TestNotifierDoesNotRecordFailedDelivery(t *testing.T) {
	store := &fakeStore{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"description":"temporary failure"}`))
	}))
	t.Cleanup(server.Close)
	bot := NewBot("token", store)
	bot.apiBase = server.URL
	notifier := &Notifier{bot: bot}
	if err := notifier.deliver(context.Background(), 1, "subscription.created", "sub-1"); err == nil {
		t.Fatal("expected delivery error")
	}
	if store.records != 0 || store.sent {
		t.Fatalf("records=%d sent=%v", store.records, store.sent)
	}
}

func TestNotifierPreventsDuplicateDelivery(t *testing.T) {
	store := &fakeStore{recordResult: true}
	bot := testBot(t, store)
	notifier := &Notifier{bot: bot}
	if err := notifier.deliver(context.Background(), 1, "subscription.created", "sub-1"); err != nil {
		t.Fatal(err)
	}
	if err := notifier.deliver(context.Background(), 1, "subscription.created", "sub-1"); err != nil {
		t.Fatal(err)
	}
	if store.records != 1 {
		t.Fatalf("records=%d", store.records)
	}
}
