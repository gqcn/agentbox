// This file implements version-one service-proxy request handlers. Every
// handler reads the authenticated AgentBox user ID from context before
// delegating to service methods that enforce Agent ownership.

package serviceproxy

import (
	"context"

	v1 "john-ai-agentbox/backend/api/serviceproxy/v1"
	"john-ai-agentbox/backend/internal/service/authctx"
	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
)

// Services lists runtime services for one visible Agent.
func (c *ControllerV1) Services(ctx context.Context, req *v1.ServicesReq) (res *v1.ServicesRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.serviceProxySvc.Services(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toRuntimeServiceListResponse(items)
	return (*v1.ServicesRes)(&out), nil
}

// Service gets one runtime service for one visible Agent.
func (c *ControllerV1) Service(ctx context.Context, req *v1.ServiceReq) (res *v1.ServiceRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.serviceProxySvc.Service(ctx, userID, req.ID, req.ServiceID)
	if err != nil {
		return nil, err
	}
	out := toRuntimeServiceResponse(item)
	return (*v1.ServiceRes)(&out), nil
}

// ServiceBridges lists explicit loopback bridges for one visible Agent.
func (c *ControllerV1) ServiceBridges(ctx context.Context, req *v1.ServiceBridgesReq) (res *v1.ServiceBridgesRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.serviceProxySvc.ServiceBridges(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	out := toBridgeListResponse(items)
	return (*v1.ServiceBridgesRes)(&out), nil
}

// CreateServiceBridge creates an explicit loopback bridge for one visible service.
func (c *ControllerV1) CreateServiceBridge(ctx context.Context, req *v1.CreateServiceBridgeReq) (res *v1.CreateServiceBridgeRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := c.serviceProxySvc.CreateServiceBridge(ctx, userID, req.ID, serviceproxysvc.BridgeInput{
		ServiceID:     req.ServiceID,
		ListenAddress: req.ListenAddress,
		Port:          req.Port,
	})
	if err != nil {
		return nil, err
	}
	out := toBridgeResponse(item)
	return (*v1.CreateServiceBridgeRes)(&out), nil
}

// DeleteServiceBridge closes one explicit loopback bridge for one visible Agent.
func (c *ControllerV1) DeleteServiceBridge(ctx context.Context, req *v1.DeleteServiceBridgeReq) (res *v1.DeleteServiceBridgeRes, err error) {
	userID, err := authctx.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}
	deleted, err := c.serviceProxySvc.DeleteServiceBridge(ctx, userID, req.ID, req.BridgeID)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteServiceBridgeRes{Deleted: deleted}, nil
}
