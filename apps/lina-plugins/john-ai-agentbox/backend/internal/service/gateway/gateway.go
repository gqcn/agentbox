// Package gateway owns AgentBox raw runtime entry points such as Chat
// WebSocket, Shell WebSocket, and TCP tunnel paths. The current migration slice
// validates AgentBox user ownership before returning structured runtime
// unavailable errors; real WebSocket and tunnel backends are migrated later.
package gateway

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	accesssvc "john-ai-agentbox/backend/internal/service/access"
)

// Service defines AgentBox raw runtime gateway behavior.
type Service interface {
	// AgentServiceHTTPProxy validates the authenticated AgentBox user and raw
	// proxy path before handing an HTTP service proxy request to the runtime.
	// The current migration slice has no key-to-service runtime scope registry,
	// so valid requests receive a structured runtime-unavailable error.
	AgentServiceHTTPProxy(ctx context.Context, userID string, escapedPath string) error
	// AgentShell validates that the authenticated AgentBox user can access the
	// target Agent before opening a runtime-backed shell WebSocket.
	AgentShell(ctx context.Context, userID string, agentID string, terminalID string, cwd string, mode string) error
	// AgentChat validates that the authenticated AgentBox user can access the
	// target Agent Chat session before opening a runtime-backed Chat WebSocket.
	AgentChat(ctx context.Context, userID string, agentID string, sessionID string, cwd string) error
	// AgentServiceTCPTunnel validates that the authenticated AgentBox user can
	// access the target Agent service before opening a runtime-backed TCP tunnel.
	AgentServiceTCPTunnel(ctx context.Context, userID string, agentID string, serviceID string, key string) error
}

// serviceImpl is the default raw gateway implementation.
type serviceImpl struct {
	accessSvc accesssvc.Service
}

var _ Service = (*serviceImpl)(nil)

// New creates a raw gateway service with explicit access dependency injection.
func New(accessSvc accesssvc.Service) (Service, error) {
	if accessSvc == nil {
		return nil, gerror.New("agentbox gateway access service is required")
	}
	return &serviceImpl{accessSvc: accessSvc}, nil
}
