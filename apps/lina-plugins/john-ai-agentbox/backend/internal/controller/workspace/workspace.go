// Package workspace implements AgentBox workspace HTTP controllers. Controllers
// resolve the current AgentBox user from request context and delegate ownership
// and runtime readiness checks to the plugin-owned workspace service.
package workspace

import (
	"github.com/gogf/gf/v2/errors/gerror"

	workspaceapi "john-ai-agentbox/backend/api/workspace"
	workspacesvc "john-ai-agentbox/backend/internal/service/workspace"
)

// ControllerV1 handles version-one AgentBox workspace APIs.
type ControllerV1 struct {
	workspaceSvc workspacesvc.Service
}

// NewV1 creates the AgentBox workspace controller.
func NewV1(workspaceSvc workspacesvc.Service) (workspaceapi.IWorkspaceV1, error) {
	if workspaceSvc == nil {
		return nil, gerror.New("agentbox workspace service is required")
	}
	return &ControllerV1{workspaceSvc: workspaceSvc}, nil
}
