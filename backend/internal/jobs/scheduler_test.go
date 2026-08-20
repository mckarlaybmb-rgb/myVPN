package jobs

import (
	"context"
	"errors"
	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	"testing"
	"time"
)

type fakeExpiry struct {
	items    []models.Subscription
	statuses int
}

func (f *fakeExpiry) ListDue(context.Context, time.Time) ([]models.Subscription, error) {
	return f.items, nil
}
func (f *fakeExpiry) UpdateStatus(_ context.Context, _ string, status string) (models.Subscription, error) {
	f.statuses++
	f.items[0].Status = status
	return f.items[0], nil
}

type fakeDisable struct {
	calls int
	fail  bool
}

func (f *fakeDisable) DisableClient(context.Context, string) error {
	f.calls++
	if f.fail {
		return errors.New("down")
	}
	return nil
}

type fakeNotify struct{ calls int }

func (f *fakeNotify) Notify(context.Context, string, string) error { f.calls++; return nil }
func TestExpirySchedulerExpiresAndNotifiesOnce(t *testing.T) {
	f := &fakeExpiry{items: []models.Subscription{{ID: "sub-1", UserID: "user-1", Status: "active", ExpiresAt: time.Now().Add(-time.Hour)}}}
	d := &fakeDisable{}
	n := &fakeNotify{}
	s := NewExpiryScheduler(f, d, n)
	s.RunOnce(context.Background())
	if f.statuses != 1 || d.calls != 1 || n.calls != 1 {
		t.Fatalf("statuses=%d disables=%d notifications=%d", f.statuses, d.calls, n.calls)
	}
	s.RunOnce(context.Background())
	if f.statuses != 1 {
		t.Fatalf("status updates=%d", f.statuses)
	}
}
func TestNodeCheckerContinuesAfterNodeFailure(t *testing.T) {
	repository := &fakeNodes{nodes: []models.Node{{ID: "1", Address: "bad", Port: 1}, {ID: "2", Address: "bad", Port: 1}}}
	NewNodeChecker(repository).RunOnce(context.Background())
	if len(repository.statuses) != 2 {
		t.Fatalf("updates=%d", len(repository.statuses))
	}
}

type fakeNodes struct {
	nodes    []models.Node
	statuses []string
}

func (f *fakeNodes) ListActive(context.Context) ([]models.Node, error) { return f.nodes, nil }
func (f *fakeNodes) UpdateNodeHealth(context.Context, string, string, *int) error {
	f.statuses = append(f.statuses, "unhealthy")
	return nil
}
