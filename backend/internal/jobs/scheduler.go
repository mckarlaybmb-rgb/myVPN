package jobs

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type ExpiryRepository interface {
	ListDue(context.Context, time.Time) ([]models.Subscription, error)
	UpdateStatus(context.Context, string, string) (models.Subscription, error)
}
type ExpiryDisabler interface {
	DisableClient(context.Context, string) error
}
type Notifier interface {
	Notify(context.Context, string, string) error
}
type ExpiryScheduler struct {
	repository ExpiryRepository
	disabler   ExpiryDisabler
	notifier   Notifier
	interval   time.Duration
}

func NewExpiryScheduler(repository ExpiryRepository, disabler ExpiryDisabler, notifier Notifier) *ExpiryScheduler {
	return &ExpiryScheduler{repository: repository, disabler: disabler, notifier: notifier, interval: time.Hour}
}
func (scheduler *ExpiryScheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(scheduler.interval)
	defer ticker.Stop()
	for {
		scheduler.RunOnce(ctx)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
func (scheduler *ExpiryScheduler) RunOnce(ctx context.Context) {
	items, err := scheduler.repository.ListDue(ctx, time.Now())
	if err != nil {
		log.Printf("expiry scheduler: %v", err)
		return
	}
	for _, item := range items {
		if item.Status == "expired" {
			continue
		}
		if item.ExpiresAt.Before(time.Now()) {
			if _, err := scheduler.repository.UpdateStatus(ctx, item.ID, "expired"); err != nil {
				log.Printf("expire subscription %s: %v", item.ID, err)
				continue
			}
			if err := scheduler.disabler.DisableClient(ctx, item.UserID); err != nil {
				log.Printf("disable expired client %s: %v", item.UserID, err)
			}
			if scheduler.notifier != nil {
				_ = scheduler.notifier.Notify(ctx, "subscription.expired", item.ID)
			}
			continue
		}
		days := int(time.Until(item.ExpiresAt).Hours() / 24)
		if days == 7 || days == 3 || days == 1 {
			if scheduler.notifier != nil {
				_ = scheduler.notifier.Notify(ctx, "subscription.expiring", item.ID)
			}
		}
	}
}

type NodeRepository interface {
	ListActive(context.Context) ([]models.Node, error)
	UpdateNodeHealth(context.Context, string, string, *int) error
}
type NodeChecker struct {
	repository NodeRepository
	interval   time.Duration
	timeout    time.Duration
}

func NewNodeChecker(repository NodeRepository) *NodeChecker {
	return &NodeChecker{repository: repository, interval: 5 * time.Minute, timeout: 5 * time.Second}
}
func (checker *NodeChecker) Run(ctx context.Context) error {
	ticker := time.NewTicker(checker.interval)
	defer ticker.Stop()
	for {
		checker.RunOnce(ctx)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
func (checker *NodeChecker) RunOnce(ctx context.Context) {
	nodes, err := checker.repository.ListActive(ctx)
	if err != nil {
		log.Printf("node health: %v", err)
		return
	}
	for _, node := range nodes {
		start := time.Now()
		connection, err := net.DialTimeout("tcp", net.JoinHostPort(node.Address, strconv.Itoa(node.Port)), checker.timeout)
		status := "healthy"
		var latency *int
		if err != nil {
			status = "unhealthy"
		} else {
			_ = connection.Close()
			value := int(time.Since(start).Milliseconds())
			latency = &value
		}
		if err := checker.repository.UpdateNodeHealth(ctx, node.ID, status, latency); err != nil {
			log.Printf("node health update %s: %v", node.ID, err)
		}
	}
}
