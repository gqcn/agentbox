// Package image implements AgentBox coding-image HTTP controllers. Controllers
// translate API DTOs and delegate all storage and safety checks to the
// plugin-owned catalog service.
package image

import (
	imageapi "john-ai-agentbox/backend/api/image"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// ControllerV1 handles version-one AgentBox coding-image APIs.
type ControllerV1 struct {
	catalogSvc catalogsvc.Service
}

// NewV1 creates the AgentBox coding-image controller.
func NewV1(catalogSvc catalogsvc.Service) imageapi.IImageV1 {
	return &ControllerV1{
		catalogSvc: catalogSvc,
	}
}
