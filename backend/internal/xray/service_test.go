package xray

import (
	"context"
	"errors"
	"testing"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
)

type fakeRepository struct {
	client  models.XrayClient
	deleted bool
	set     *bool
}

func (repository *fakeRepository) Create(_ context.Context, client models.XrayClient) (models.XrayClient, error) {
	repository.client = client
	repository.client.ID = "client-1"
	return repository.client, nil
}
func (repository *fakeRepository) GetByUser(context.Context, string) (models.XrayClient, error) {
	if repository.client.ID == "" {
		return models.XrayClient{}, errors.New("missing client")
	}
	return repository.client, nil
}
func (repository *fakeRepository) Delete(context.Context, string) error {
	repository.deleted = true
	return nil
}
func (repository *fakeRepository) SetEnabled(_ context.Context, _ string, enabled bool) error {
	repository.set = &enabled
	repository.client.Enabled = enabled
	return nil
}

type fakeRuntime struct {
	created, deleted, enabled, disabled int
	failCreate                          bool
}

func (runtime *fakeRuntime) CreateClient(context.Context, models.XrayClient) error {
	runtime.created++
	if runtime.failCreate {
		return errors.New("runtime unavailable")
	}
	return nil
}
func (runtime *fakeRuntime) DeleteClient(context.Context, models.XrayClient) error {
	runtime.deleted++
	return nil
}
func (runtime *fakeRuntime) EnableClient(context.Context, models.XrayClient) error {
	runtime.enabled++
	return nil
}
func (runtime *fakeRuntime) DisableClient(context.Context, models.XrayClient) error {
	runtime.disabled++
	return nil
}

func TestServiceCreatesVLESSClientAndRollsBackRuntimeFailure(t *testing.T) {
	repository := &fakeRepository{}
	runtime := &fakeRuntime{}
	service := NewService(repository, runtime, "vless-inbound")
	client, err := service.CreateClient(context.Background(), "user-1", "user@example.com")
	if err != nil || client.Protocol != "vless" || client.UUID == "" || client.Config["inbound_tag"] != "vless-inbound" {
		t.Fatalf("created client=%#v err=%v", client, err)
	}
	if runtime.created != 1 {
		t.Fatalf("runtime create calls=%d", runtime.created)
	}

	failingRuntime := &fakeRuntime{failCreate: true}
	failingRepository := &fakeRepository{}
	if _, err := NewService(failingRepository, failingRuntime, "tag").CreateClient(context.Background(), "user-2", "two@example.com"); err == nil || !failingRepository.deleted {
		t.Fatalf("expected provisioning failure with rollback: err=%v deleted=%v", err, failingRepository.deleted)
	}
}

func TestServiceDisablesAndEnablesClient(t *testing.T) {
	repository := &fakeRepository{client: models.XrayClient{ID: "client-1", Enabled: true}}
	runtime := &fakeRuntime{}
	service := NewService(repository, runtime, "tag")
	if err := service.DisableClient(context.Background(), "user-1"); err != nil {
		t.Fatal(err)
	}
	if err := service.EnableClient(context.Background(), "user-1"); err != nil {
		t.Fatal(err)
	}
	if runtime.disabled != 1 || runtime.enabled != 1 || repository.set == nil || !*repository.set {
		t.Fatalf("runtime=%#v repository=%#v", runtime, repository)
	}
}
