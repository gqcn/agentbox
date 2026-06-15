// This file defines the public service-proxy API contract surface for the
// AgentBox plugin. Versioned DTOs stay in subpackages while this package keeps
// the controller interface used by source-plugin route registration.

package serviceproxy

import (
	"context"

	v1 "john-ai-agentbox/backend/api/serviceproxy/v1"
)

// IServiceProxyV1 defines AgentBox runtime service discovery and bridge handlers.
type IServiceProxyV1 interface {
	// Services lists runtime services for one authenticated-user-owned Agent.
	Services(ctx context.Context, req *v1.ServicesReq) (res *v1.ServicesRes, err error)
	// Service gets one runtime service for one authenticated-user-owned Agent.
	Service(ctx context.Context, req *v1.ServiceReq) (res *v1.ServiceRes, err error)
	// ServiceBridges lists explicit loopback bridges for one authenticated-user-owned Agent.
	ServiceBridges(ctx context.Context, req *v1.ServiceBridgesReq) (res *v1.ServiceBridgesRes, err error)
	// CreateServiceBridge creates an explicit loopback bridge for one visible service.
	CreateServiceBridge(ctx context.Context, req *v1.CreateServiceBridgeReq) (res *v1.CreateServiceBridgeRes, err error)
	// DeleteServiceBridge closes one explicit loopback bridge for one visible Agent.
	DeleteServiceBridge(ctx context.Context, req *v1.DeleteServiceBridgeReq) (res *v1.DeleteServiceBridgeRes, err error)
}
