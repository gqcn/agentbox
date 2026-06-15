// Package container implements AgentBox container runtime HTTP controllers.
// Controllers resolve the current AgentBox user from request context and
// delegate runtime readiness checks to the plugin-owned container service.
package container

import (
	"github.com/gogf/gf/v2/errors/gerror"

	containerapi "john-ai-agentbox/backend/api/container"
	containersvc "john-ai-agentbox/backend/internal/service/container"
)

// ControllerV1 handles version-one AgentBox container APIs.
type ControllerV1 struct {
	containerSvc containersvc.Service
}

// NewV1 creates the AgentBox container controller.
func NewV1(containerSvc containersvc.Service) (containerapi.IContainerV1, error) {
	if containerSvc == nil {
		return nil, gerror.New("agentbox container service is required")
	}
	return &ControllerV1{containerSvc: containerSvc}, nil
}
