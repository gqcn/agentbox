// Package prompt implements AgentBox prompt-template HTTP controllers.
// Controllers delegate registry, persistence, and rendering behavior to the
// plugin-owned prompt service.
package prompt

import (
	promptapi "john-ai-agentbox/backend/api/prompt"
	promptsvc "john-ai-agentbox/backend/internal/service/prompt"
)

// ControllerV1 handles version-one AgentBox prompt APIs.
type ControllerV1 struct {
	promptSvc promptsvc.Service
}

// NewV1 creates the AgentBox prompt controller.
func NewV1(promptSvc promptsvc.Service) promptapi.IPromptV1 {
	return &ControllerV1{
		promptSvc: promptSvc,
	}
}
