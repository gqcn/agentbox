// Package serviceproxy implements AgentBox runtime service-proxy HTTP
// controllers. Controllers resolve the current AgentBox user from request
// context and delegate ownership and runtime readiness checks to the plugin
// service-proxy service.
package serviceproxy

import (
	"github.com/gogf/gf/v2/errors/gerror"

	serviceproxyapi "john-ai-agentbox/backend/api/serviceproxy"
	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
)

// ControllerV1 handles version-one AgentBox runtime service-proxy APIs.
type ControllerV1 struct {
	serviceProxySvc serviceproxysvc.Service
}

// NewV1 creates the AgentBox service-proxy controller.
func NewV1(serviceProxySvc serviceproxysvc.Service) (serviceproxyapi.IServiceProxyV1, error) {
	if serviceProxySvc == nil {
		return nil, gerror.New("agentbox service proxy service is required")
	}
	return &ControllerV1{serviceProxySvc: serviceProxySvc}, nil
}
