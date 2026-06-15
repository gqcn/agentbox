// This file maps AgentBox setting service projections to public DTOs. It keeps
// generated entities and service-internal fields out of HTTP responses.

package setting

import (
	v1 "john-ai-agentbox/backend/api/setting/v1"
	settingsvc "john-ai-agentbox/backend/internal/service/setting"
)

func toSettingResponse(item settingsvc.SettingInfo) v1.SettingInfo {
	return v1.SettingInfo{
		Key:       item.Key,
		Value:     item.Value,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
