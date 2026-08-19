package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/mckarlaybmb-rgb/myVPN/backend/internal/models"
	command "github.com/xtls/xray-core/app/proxyman/command"
	protocol "github.com/xtls/xray-core/common/protocol"
	serial "github.com/xtls/xray-core/common/serial"
	vless "github.com/xtls/xray-core/proxy/vless"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Runtime interface {
	CreateClient(context.Context, models.XrayClient) error
	DeleteClient(context.Context, models.XrayClient) error
	EnableClient(context.Context, models.XrayClient) error
	DisableClient(context.Context, models.XrayClient) error
}

type HandlerServiceRuntime struct {
	address    string
	inboundTag string
}

func NewHandlerServiceRuntime(address, inboundTag string) *HandlerServiceRuntime {
	return &HandlerServiceRuntime{address: address, inboundTag: inboundTag}
}

func (runtime *HandlerServiceRuntime) withClient(ctx context.Context, operation *serial.TypedMessage) error {
	if runtime.address == "" || runtime.inboundTag == "" {
		return fmt.Errorf("xray API address and inbound tag are required")
	}
	connection, err := grpc.DialContext(ctx, runtime.address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, "tcp", address)
	}))
	if err != nil {
		return err
	}
	defer connection.Close()
	_, err = command.NewHandlerServiceClient(connection).AlterInbound(ctx, &command.AlterInboundRequest{Tag: runtime.inboundTag, Operation: operation})
	return err
}

func (runtime *HandlerServiceRuntime) CreateClient(ctx context.Context, client models.XrayClient) error {
	return runtime.withClient(ctx, serial.ToTypedMessage(&command.AddUserOperation{User: runtime.user(client)}))
}

func (runtime *HandlerServiceRuntime) DeleteClient(ctx context.Context, client models.XrayClient) error {
	return runtime.withClient(ctx, serial.ToTypedMessage(&command.RemoveUserOperation{Email: client.Email}))
}

func (runtime *HandlerServiceRuntime) EnableClient(ctx context.Context, client models.XrayClient) error {
	return runtime.CreateClient(ctx, client)
}

func (runtime *HandlerServiceRuntime) DisableClient(ctx context.Context, client models.XrayClient) error {
	return runtime.DeleteClient(ctx, client)
}

func (runtime *HandlerServiceRuntime) user(client models.XrayClient) *protocol.User {
	flow := ""
	if value, ok := client.Config["flow"].(string); ok {
		flow = value
	}
	return &protocol.User{Email: client.Email, Account: serial.ToTypedMessage(&vless.Account{Id: client.UUID, Flow: flow, Encryption: "none"})}
}

func BuildVLESSConfig(inboundTag string) map[string]any {
	return map[string]any{"protocol": "vless", "inbound_tag": inboundTag, "flow": "xtls-rprx-vision", "encryption": "none"}
}

func ConfigJSON(config map[string]any) ([]byte, error) { return json.Marshal(config) }
