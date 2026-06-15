// Package setting implements AgentBox user-scoped setting HTTP controllers.
// Controllers resolve the current AgentBox user from request context and
// delegate persistence to the plugin-owned setting service.
package setting

import (
	settingapi "john-ai-agentbox/backend/api/setting"
	settingsvc "john-ai-agentbox/backend/internal/service/setting"
)

// ControllerV1 handles version-one AgentBox setting APIs.
type ControllerV1 struct {
	settingSvc settingsvc.Service
}

// NewV1 creates the AgentBox setting controller.
func NewV1(settingSvc settingsvc.Service) settingapi.ISettingV1 {
	return &ControllerV1{
		settingSvc: settingSvc,
	}
}
