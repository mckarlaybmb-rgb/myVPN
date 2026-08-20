package xray

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type Repository interface {
	Create(context.Context, models.XrayClient) (models.XrayClient, error)
	GetByUser(context.Context, string) (models.XrayClient, error)
	Delete(context.Context, string) error
	SetEnabled(context.Context, string, bool) error
}

type Service struct {
	repository Repository
	runtime    Runtime
	inboundTag string
}

func NewService(repository Repository, runtime Runtime, inboundTag string) *Service {
	return &Service{repository: repository, runtime: runtime, inboundTag: inboundTag}
}

func (service *Service) CreateClient(ctx context.Context, userID, email string) (models.XrayClient, error) {
	client := models.XrayClient{UserID: userID, Email: email, Protocol: "vless", Config: BuildVLESSConfig(service.inboundTag), Enabled: true}
	var err error
	if client.UUID, err = newUUID(); err != nil {
		return models.XrayClient{}, fmt.Errorf("generate xray client UUID: %w", err)
	}
	client.SubscriptionURL, err = service.runtime.AddUser(ctx, client)
	if err != nil {
		return models.XrayClient{}, fmt.Errorf("provision x-ui client: %w", err)
	}
	created, err := service.repository.Create(ctx, client)
	if err != nil {
		_ = service.runtime.RemoveUser(ctx, client)
		return models.XrayClient{}, fmt.Errorf("persist xray client: %w", err)
	}
	return created, nil
}
func (service *Service) DeleteClient(ctx context.Context, userID string) error {
	client, err := service.repository.GetByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("load xray client: %w", err)
	}
	if err := service.runtime.RemoveUser(ctx, client); err != nil {
		return fmt.Errorf("delete xray client from x-ui: %w", err)
	}
	if err := service.repository.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete xray client record: %w", err)
	}
	return nil
}
func (service *Service) EnableClient(ctx context.Context, userID string) error {
	client, err := service.repository.GetByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("load xray client: %w", err)
	}
	if client.Enabled {
		return nil
	}
	if err := service.runtime.EnableClient(ctx, client); err != nil {
		return fmt.Errorf("enable xray client: %w", err)
	}
	return service.repository.SetEnabled(ctx, userID, true)
}
func (service *Service) DisableClient(ctx context.Context, userID string) error {
	client, err := service.repository.GetByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("load xray client: %w", err)
	}
	if !client.Enabled {
		return nil
	}
	if err := service.runtime.DisableClient(ctx, client); err != nil {
		return fmt.Errorf("disable xray client: %w", err)
	}
	return service.repository.SetEnabled(ctx, userID, false)
}
func newUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}
