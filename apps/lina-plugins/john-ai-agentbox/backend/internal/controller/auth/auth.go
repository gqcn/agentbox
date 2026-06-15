// Package auth implements AgentBox authentication HTTP controllers. These
// controllers are plugin-owned and use the plugin's independent session state,
// not LinaPro management-workbench authentication.
package auth

import (
	authapi "john-ai-agentbox/backend/api/auth"
	authsvc "john-ai-agentbox/backend/internal/service/auth"
)

// ControllerV1 handles version-one AgentBox authentication APIs.
type ControllerV1 struct {
	authSvc authsvc.Service
}

// NewV1 creates the AgentBox authentication controller.
func NewV1(authSvc authsvc.Service) authapi.IAuthV1 {
	return &ControllerV1{
		authSvc: authSvc,
	}
}
