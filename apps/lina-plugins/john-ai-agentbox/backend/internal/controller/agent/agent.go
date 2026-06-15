// Package agent implements AgentBox coding-agent HTTP controllers. Controllers
// read the authenticated AgentBox user from context and delegate all
// user/resource ownership checks to the plugin-owned catalog service.
package agent

import (
	"github.com/gogf/gf/v2/errors/gerror"

	agentapi "john-ai-agentbox/backend/api/agent"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
	chatsvc "john-ai-agentbox/backend/internal/service/chat"
	terminalsvc "john-ai-agentbox/backend/internal/service/terminal"
)

// ControllerV1 handles version-one AgentBox coding-agent APIs.
type ControllerV1 struct {
	catalogSvc  catalogsvc.Service
	chatSvc     chatsvc.Service
	terminalSvc terminalsvc.Service
}

// NewV1 creates the AgentBox coding-agent controller.
func NewV1(catalogSvc catalogsvc.Service, chatSvc chatsvc.Service, terminalSvc terminalsvc.Service) (agentapi.IAgentV1, error) {
	if catalogSvc == nil {
		return nil, gerror.New("agentbox catalog service is required")
	}
	if chatSvc == nil {
		return nil, gerror.New("agentbox chat service is required")
	}
	if terminalSvc == nil {
		return nil, gerror.New("agentbox terminal service is required")
	}
	return &ControllerV1{
		catalogSvc:  catalogSvc,
		chatSvc:     chatSvc,
		terminalSvc: terminalSvc,
	}, nil
}
