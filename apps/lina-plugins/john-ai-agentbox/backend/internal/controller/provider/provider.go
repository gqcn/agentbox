// Package provider implements AgentBox provider HTTP controllers. Controllers
// only translate versioned API DTOs and delegate storage behavior to the
// plugin-owned catalog service.
package provider

import (
	providerapi "john-ai-agentbox/backend/api/provider"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// ControllerV1 handles version-one AgentBox provider APIs.
type ControllerV1 struct {
	catalogSvc catalogsvc.Service
}

// NewV1 creates the AgentBox provider controller.
func NewV1(catalogSvc catalogsvc.Service) providerapi.IProviderV1 {
	return &ControllerV1{
		catalogSvc: catalogSvc,
	}
}
