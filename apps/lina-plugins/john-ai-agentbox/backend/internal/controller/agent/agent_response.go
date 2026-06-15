// This file converts AgentBox catalog agent projections into versioned API
// DTOs without exposing service internals through controller methods.

package agent

import (
	v1 "john-ai-agentbox/backend/api/agent/v1"
	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

// toAgentResponse maps a service agent projection into a public DTO.
func toAgentResponse(item catalogsvc.AgentInfo) v1.AgentInfo {
	return v1.AgentInfo{
		ID:             item.ID,
		UserID:         item.UserID,
		Name:           item.Name,
		ProviderID:     item.ProviderID,
		ProviderName:   item.ProviderName,
		ModelName:      item.ModelName,
		ModelProtocol:  item.ModelProtocol,
		ImageID:        item.ImageID,
		ImageName:      item.ImageName,
		ImageRef:       item.ImageRef,
		AgentType:      item.AgentType,
		IconKey:        item.IconKey,
		Notes:          item.Notes,
		RuntimeStatus:  item.RuntimeStatus,
		ActivityStatus: item.ActivityStatus,
		ContainerID:    item.ContainerID,
		DockerID:       item.DockerID,
		DeletedAt:      item.DeletedAt,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

// toAgentListResponse maps service agent projections into public DTOs.
func toAgentListResponse(items []catalogsvc.AgentInfo) []v1.AgentInfo {
	result := make([]v1.AgentInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toAgentResponse(item))
	}
	return result
}

// toChangeImageResponse maps a service image-switch result into a public DTO.
func toChangeImageResponse(out *catalogsvc.ChangeAgentImageOutput) *v1.ChangeAgentImageResponse {
	if out == nil {
		return &v1.ChangeAgentImageResponse{}
	}
	return &v1.ChangeAgentImageResponse{
		Agent:          toAgentResponse(out.Agent),
		LostPaths:      append([]string(nil), out.LostPaths...),
		PreservedPaths: append([]string(nil), out.PreservedPaths...),
	}
}
