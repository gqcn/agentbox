// Package ai implements AgentBox AI capability HTTP controllers. Controllers
// delegate all plugin-owned AI persistence and provider test behavior to the
// AI service.
package ai

import (
	aiapi "john-ai-agentbox/backend/api/ai"
	aisvc "john-ai-agentbox/backend/internal/service/ai"
)

// ControllerV1 handles version-one AgentBox AI capability APIs.
type ControllerV1 struct {
	aiSvc aisvc.Service
}

// NewV1 creates the AgentBox AI capability controller.
func NewV1(aiSvc aisvc.Service) aiapi.IAiV1 {
	return &ControllerV1{
		aiSvc: aiSvc,
	}
}
